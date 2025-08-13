package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func normalizeAndValidateRelationship(in string) (bool, string) {
	s := strings.TrimSpace(strings.ToUpper(in))
	switch s {
	case "REFERENCES":
		return true, "REFERENCES"
	case "CONTAINS":
		return true, "CONTAINS"
	case "RELATED":
		return true, "RELATED"
	case "FOLLOWS":
		return true, "FOLLOWS"
	default:
		return false, ""
	}
}

type Neo4jRepository struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jRepository(driver neo4j.DriverWithContext) domain.GraphRepository {
	return &Neo4jRepository{driver: driver}
}

var _ domain.GraphRepository = (*Neo4jRepository)(nil)

// Content Node Operations
func (r *Neo4jRepository) CreateContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, runErr := tx.Run(ctx,
			"MERGE (c:Content {contentId: $contentId}) SET c.spaceId = $spaceId",
			map[string]any{
				"contentId": contentID.String(),
				"spaceId":   spaceID.String(),
			},
		)
		return nil, runErr
	})
	return err
}

func (r *Neo4jRepository) DeleteContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Delete relationships, then node scoped by spaceId
		_, runErr := tx.Run(ctx,
			"MATCH (c:Content {contentId: $id, spaceId: $spaceId})-[r]-() DELETE r",
			map[string]any{"id": contentID.String(), "spaceId": spaceID.String()},
		)
		if runErr != nil {
			return nil, runErr
		}
		_, runErr = tx.Run(ctx,
			"MATCH (c:Content {contentId: $id, spaceId: $spaceId}) DELETE c",
			map[string]any{"id": contentID.String(), "spaceId": spaceID.String()},
		)
		return nil, runErr
	})
	return err
}

func (r *Neo4jRepository) CreateLink(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, runErr := tx.Run(ctx,
			"MATCH (a:Content {contentId: $from}), (b:Content {contentId: $to}) MERGE (a)-[:LINKS_TO]->(b)",
			map[string]any{"from": fromContentID.String(), "to": toContentID.String()},
		)
		return nil, runErr
	})
	return err
}

func (r *Neo4jRepository) GetBacklinks(ctx context.Context, contentID uuid.UUID) ([]uuid.UUID, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		records, runErr := tx.Run(ctx,
			"MATCH (a:Content)-[:LINKS_TO]->(b:Content {contentId: $id}) RETURN a.contentId AS id",
			map[string]any{"id": contentID.String()},
		)
		if runErr != nil {
			return nil, runErr
		}
		var ids []uuid.UUID
		for records.Next(ctx) {
			val, _ := records.Record().Get("id")
			if s, ok := val.(string); ok {
				if u, parseErr := uuid.Parse(s); parseErr == nil {
					ids = append(ids, u)
				}
			}
		}
		return ids, records.Err()
	})
	if err != nil {
		return nil, err
	}
	return result.([]uuid.UUID), nil
}
