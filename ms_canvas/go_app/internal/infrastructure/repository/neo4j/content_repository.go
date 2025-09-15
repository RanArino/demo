package neo4j

import (
	"context"
	"time"

	"demo/ms_canvas/go_app/internal/domain"
	"github.com/google/uuid"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type ContentRepo struct {
	driver *Driver
}

func NewContentRepo(driver *Driver) *ContentRepo { return &ContentRepo{driver: driver} }

func (r *ContentRepo) CreateContentNode(node domain.ContentNode) error {
	ctx := context.Background()
	sess := r.driver.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{
			"content_source_id": node.ContentSourceID.String(),
			"now":               time.Now().UTC().Format(time.RFC3339),
			"id":                uuid.New().String(),
		}
		_, err := tx.Run(ctx, `
			MERGE (n:ContentNode {content_source_id: $content_source_id})
			ON CREATE SET n.id = $id, n.created_at = datetime($now), n.updated_at = datetime($now)
			ON MATCH SET n.updated_at = datetime($now)
		`, params)
		return nil, err
	})
	return err
}
