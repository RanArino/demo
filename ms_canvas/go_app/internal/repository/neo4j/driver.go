package neo4j

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Driver struct {
	driver            neo.DriverWithContext
	dbName            string
	vectorDim         int
	constraintTimeout time.Duration
	indexTimeout      time.Duration
	maxEnsureRetries  int
	ensureRetryDelay  time.Duration
}

// DriverOptions contains optional settings for the Neo4j driver
type DriverOptions struct {
	VectorDimensions  int
	ConstraintTimeout time.Duration
	IndexTimeout      time.Duration
	MaxEnsureRetries  int
	EnsureRetryDelay  time.Duration
}

// NewDriver creates a new Driver. Pass nil for opts to use defaults.
func NewDriver(uri, user, pass, db string, opts *DriverOptions) (*Driver, error) {
	// Basic early validation: attempt to parse URI and resolve the host to give a clearer error
	if u, err := url.Parse(uri); err == nil {
		host := u.Host
		if strings.Contains(host, ":") {
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}
		}
		if host != "" {
			if err := retryLookupHost(host); err != nil {
				return nil, fmt.Errorf("unable to resolve neo4j host %q from URI %q: %w", host, uri, err)
			}
		}
	}

	drv, err := neo.NewDriverWithContext(uri, neo.BasicAuth(user, pass, ""))
	if err != nil {
		return nil, err
	}

	// Default values
	vectorDim := 1536
	if opts != nil && opts.VectorDimensions > 0 {
		vectorDim = opts.VectorDimensions
	}

	constraintTimeout := 15 * time.Second
	indexTimeout := 30 * time.Second
	maxEnsureRetries := 2
	ensureRetryDelay := 5 * time.Second

	if opts != nil {
		if opts.ConstraintTimeout > 0 {
			constraintTimeout = opts.ConstraintTimeout
		}
		if opts.IndexTimeout > 0 {
			indexTimeout = opts.IndexTimeout
		}
		if opts.MaxEnsureRetries > 0 {
			maxEnsureRetries = opts.MaxEnsureRetries
		}
		if opts.EnsureRetryDelay > 0 {
			ensureRetryDelay = opts.EnsureRetryDelay
		}
	}

	return &Driver{
		driver:            drv,
		dbName:            db,
		vectorDim:         vectorDim,
		constraintTimeout: constraintTimeout,
		indexTimeout:      indexTimeout,
		maxEnsureRetries:  maxEnsureRetries,
		ensureRetryDelay:  ensureRetryDelay,
	}, nil
}

func (d *Driver) Close(ctx context.Context) error {
	return d.driver.Close(ctx)
}

// NewSession delegates session creation to the underlying Neo4j driver.
func (d *Driver) NewSession(ctx context.Context, config neo.SessionConfig) neo.SessionWithContext {
	return d.driver.NewSession(ctx, config)
}

func (d *Driver) EnsureConstraints(ctx context.Context) error {
	return d.ensureWithRetry(ctx, d.constraintTimeout, func(runCtx context.Context, tx neo.ManagedTransaction) error {
		_, err := tx.Run(runCtx, "CREATE CONSTRAINT content_source_unique IF NOT EXISTS FOR (n:ContentNode) REQUIRE n.content_source_id IS UNIQUE", nil)
		return err
	})
}

// EnsureIndexes creates helpful BTREE and, if supported, vector indexes.
func (d *Driver) EnsureIndexes(ctx context.Context) error {
	return d.ensureWithRetry(ctx, d.indexTimeout, func(runCtx context.Context, tx neo.ManagedTransaction) error {
		if _, err := tx.Run(runCtx, "CREATE INDEX contentnode_space IF NOT EXISTS FOR (n:ContentNode) ON (n.space_id)", nil); err != nil {
			return err
		}
		if _, err := tx.Run(runCtx, "CREATE INDEX chunknode_csid IF NOT EXISTS FOR (n:ChunkNode) ON (n.content_source_id)", nil); err != nil {
			log.Printf("[Neo4j] create index chunknode_csid: %v", err)
		}

		vectorIndexQueries := []string{
			fmt.Sprintf("CREATE VECTOR INDEX clusternode_embedding IF NOT EXISTS FOR (n:ClusterNode) ON (n.embedding) OPTIONS {indexConfig: {`vector.dimensions`: %d, `vector.similarity_function`: 'cosine'}}", d.vectorDim),
			fmt.Sprintf("CREATE VECTOR INDEX contentnode_embedding IF NOT EXISTS FOR (n:ContentNode) ON (n.embedding) OPTIONS {indexConfig: {`vector.dimensions`: %d, `vector.similarity_function`: 'cosine'}}", d.vectorDim),
			fmt.Sprintf("CREATE VECTOR INDEX chunknode_embedding IF NOT EXISTS FOR (n:ChunkNode) ON (n.embedding) OPTIONS {indexConfig: {`vector.dimensions`: %d, `vector.similarity_function`: 'cosine'}}", d.vectorDim),
		}

		for _, query := range vectorIndexQueries {
			if _, err := tx.Run(runCtx, query, nil); err != nil {
				log.Printf("[Neo4j] create vector index (embedding): %v", err)
			}
		}
		return nil
	})
}

func (d *Driver) ensureWithRetry(ctx context.Context, timeout time.Duration, work func(context.Context, neo.ManagedTransaction) error) error {
	attempt := 0
	for {
		attempt++
		ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
		sess := d.driver.NewSession(ctxTimeout, neo.SessionConfig{DatabaseName: d.dbName})
		err := func() error {
			defer cancel()
			defer sess.Close(ctxTimeout)
			_, err := sess.ExecuteWrite(ctxTimeout, func(tx neo.ManagedTransaction) (any, error) {
				if err := work(ctxTimeout, tx); err != nil {
					return nil, err
				}
				return nil, nil
			})
			return err
		}()
		if err == nil {
			return nil
		}

		log.Printf("[Neo4j] ensure attempt %d failed: %v", attempt, err)
		if attempt >= d.maxEnsureRetries {
			return err
		}

		delay := d.ensureRetryDelay
		if delay <= 0 {
			delay = 2 * time.Second
		}
		select {
		case <-ctx.Done():
			return errors.Join(err, ctx.Err())
		case <-time.After(delay):
		}
	}
}

func retryLookupHost(host string) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		if _, err := net.LookupHost(host); err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
		} else {
			return nil
		}
	}
	return lastErr
}
