package neo4j

import (
	"context"
	"log"

	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Driver struct {
	driver  neo.DriverWithContext
	dbName  string
}

func NewDriver(uri, user, pass, db string) (*Driver, error) {
	drv, err := neo.NewDriverWithContext(uri, neo.BasicAuth(user, pass, ""))
	if err != nil {
		return nil, err
	}
	return &Driver{driver: drv, dbName: db}, nil
}

func (d *Driver) Close(ctx context.Context) error {
	return d.driver.Close(ctx)
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
