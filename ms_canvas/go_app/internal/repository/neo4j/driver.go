package neo4j

import (
	"context"
	"fmt"
	"log"

	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Driver struct {
	driver    neo.DriverWithContext
	dbName    string
	vectorDim int
}

// DriverOptions contains optional settings for the Neo4j driver
type DriverOptions struct {
	VectorDimensions int
}

// NewDriver creates a new Driver. Pass nil for opts to use defaults.
func NewDriver(uri, user, pass, db string, opts *DriverOptions) (*Driver, error) {
	drv, err := neo.NewDriverWithContext(uri, neo.BasicAuth(user, pass, ""))
	if err != nil {
		return nil, err
	}

	// Default values
	vectorDim := 1536
	if opts != nil && opts.VectorDimensions > 0 {
		vectorDim = opts.VectorDimensions
	}

	return &Driver{driver: drv, dbName: db, vectorDim: vectorDim}, nil
}

func (d *Driver) Close(ctx context.Context) error {
	return d.driver.Close(ctx)
}

// NewSession delegates session creation to the underlying Neo4j driver.
func (d *Driver) NewSession(ctx context.Context, config neo.SessionConfig) neo.SessionWithContext {
	return d.driver.NewSession(ctx, config)
}

func (d *Driver) EnsureConstraints(ctx context.Context) error {
	sess := d.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: d.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, "CREATE CONSTRAINT content_source_unique IF NOT EXISTS FOR (n:ContentNode) REQUIRE n.content_source_id IS UNIQUE", nil)
		return nil, err
	})
	if err != nil {
		log.Printf("[Neo4j] ensure constraints error: %v", err)
	}
	return err
}

// EnsureIndexes creates helpful BTREE and, if supported, vector indexes.
func (d *Driver) EnsureIndexes(ctx context.Context) error {
	sess := d.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: d.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		// BTREE indexes for common lookups
		if _, e := tx.Run(ctx, "CREATE INDEX contentnode_space IF NOT EXISTS FOR (n:ContentNode) ON (n.space_id)", nil); e != nil {
			log.Printf("[Neo4j] create index contentnode_space: %v", e)
		}
		if _, e := tx.Run(ctx, "CREATE INDEX chunknode_csid IF NOT EXISTS FOR (n:ChunkNode) ON (n.content_source_id)", nil); e != nil {
			log.Printf("[Neo4j] create index chunknode_csid: %v", e)
		}

		// Vector index (Neo4j 5.11+). This may fail on older versions; log and continue.
		query := fmt.Sprintf("CREATE VECTOR INDEX chunknode_embedding IF NOT EXISTS FOR (n:ChunkNode) ON (n.embedding) WITH {indexConfig: {`vector.dimensions`: %d, `vector.similarity_function`: 'cosine'}}", d.vectorDim)
		if _, e := tx.Run(ctx, query, nil); e != nil {
			log.Printf("[Neo4j] create vector index (embedding): %v", e)
		}
		return nil, nil
	})
	return err
}
