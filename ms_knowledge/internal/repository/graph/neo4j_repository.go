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
func (r *Neo4jRepository) CreateContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID, title string, contentSummary *string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		params := map[string]any{
			"contentId": contentID.String(),
			"spaceId":   spaceID.String(),
			"title":     title,
		}
		setSummary := ""
		if contentSummary != nil {
			params["summary"] = *contentSummary
			setSummary = ", c.contentSummary = $summary"
		}
		query := "MERGE (c:Content {contentId: $contentId}) SET c.spaceId = $spaceId, c.title = $title" + setSummary
		_, runErr := tx.Run(ctx, query, params)
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

// Knowledge Link Operations (CRUD)
func (r *Neo4jRepository) CreateKnowledgeLink(ctx context.Context, link *domain.KnowledgeLink) error {
	if link.FromContentID == link.ToContentID {
		return fmt.Errorf("cannot create self-link")
	}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Use the relation type as the actual Neo4j relationship label (validated & normalized)
		valid, relType := normalizeAndValidateRelationship(string(link.RelationType))
		if !valid {
			return nil, fmt.Errorf("invalid relationship type: %s", link.RelationType)
		}

		query := fmt.Sprintf(`
		    MATCH (a:Content {contentId: $from}), (b:Content {contentId: $to})
		    WHERE a.spaceId = b.spaceId
		    MERGE (a)-[r:%s {
		        linkId: $linkId,
		        weight: $weight,
		        createdAt: $createdAt,
		        updatedAt: $updatedAt
		    }]->(b)
		    RETURN r.linkId as linkId
		`, relType)

		var weight float64
		if link.Weight != nil {
			weight = *link.Weight
		} else {
			weight = 1.0
		}

		result, runErr := tx.Run(ctx, query, map[string]any{
			"from":      link.FromContentID.String(),
			"to":        link.ToContentID.String(),
			"linkId":    link.ID.String(),
			"weight":    weight,
			"createdAt": link.CreatedAt.Unix(),
			"updatedAt": link.UpdatedAt.Unix(),
		})

		if runErr != nil {
			return nil, runErr
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}
		return nil, fmt.Errorf("failed to create link")
	})
	return err
}

func (r *Neo4jRepository) GetKnowledgeLink(ctx context.Context, id uuid.UUID) (*domain.EnrichedKnowledgeLink, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (a:Content)-[r {linkId: $linkId}]->(b:Content)
			RETURN a.contentId as fromId, a.title as fromTitle, a.contentSummary as fromSummary,
			       b.contentId as toId,   b.title as toTitle,   b.contentSummary as toSummary,
			       type(r) as relationType, r.weight as weight, r.createdAt as createdAt, r.updatedAt as updatedAt
		`

		records, runErr := tx.Run(ctx, query, map[string]any{"linkId": id.String()})
		if runErr != nil {
			return nil, runErr
		}

		if records.Next(ctx) {
			record := records.Record()
			fromID, _ := uuid.Parse(record.Values[0].(string))
			fromTitle, _ := record.Values[1].(string)
			var fromSummary *string
			if v := record.Values[2]; v != nil {
				if s, ok := v.(string); ok {
					fromSummary = &s
				}
			}
			toID, _ := uuid.Parse(record.Values[3].(string))
			toTitle, _ := record.Values[4].(string)
			var toSummary *string
			if v := record.Values[5]; v != nil {
				if s, ok := v.(string); ok {
					toSummary = &s
				}
			}
			relationType := domain.RelationType(record.Values[6].(string))
			weight := record.Values[7].(float64)
			createdAt := time.Unix(record.Values[8].(int64), 0)
			updatedAt := time.Unix(record.Values[9].(int64), 0)

			return &domain.EnrichedKnowledgeLink{
				ID:           id,
				From:         domain.ContentPreview{ID: fromID, Title: fromTitle, ContentSummary: fromSummary},
				To:           domain.ContentPreview{ID: toID, Title: toTitle, ContentSummary: toSummary},
				RelationType: relationType,
				Weight:       &weight,
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			}, nil
		}
		return nil, fmt.Errorf("link not found")
	})

	if err != nil {
		return nil, err
	}
	return result.(*domain.EnrichedKnowledgeLink), nil
}

func (r *Neo4jRepository) ListKnowledgeLinks(ctx context.Context, filter domain.LinkFilter) ([]*domain.EnrichedKnowledgeLink, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var query string
		params := map[string]any{"contentId": filter.ContentID.String()}

		switch filter.Direction {
		case domain.LinkDirectionInbound:
			query = `
				MATCH (a:Content)-[r]->(b:Content {contentId: $contentId})
				RETURN a.contentId as fromId, a.title as fromTitle, a.contentSummary as fromSummary,
				       b.contentId as toId,   b.title as toTitle,   b.contentSummary as toSummary,
				       r.linkId as linkId, type(r) as relationType, r.weight as weight,
				       r.createdAt as createdAt, r.updatedAt as updatedAt
				ORDER BY r.createdAt DESC
			`
		case domain.LinkDirectionOutbound:
			query = `
				MATCH (a:Content {contentId: $contentId})-[r]->(b:Content)
				RETURN a.contentId as fromId, a.title as fromTitle, a.contentSummary as fromSummary,
				       b.contentId as toId,   b.title as toTitle,   b.contentSummary as toSummary,
				       r.linkId as linkId, type(r) as relationType, r.weight as weight,
				       r.createdAt as createdAt, r.updatedAt as updatedAt
				ORDER BY r.createdAt DESC
			`
		case domain.LinkDirectionBoth:
			query = `
				MATCH (a:Content)-[r]-(b:Content)
				WHERE a.contentId = $contentId OR b.contentId = $contentId
				RETURN a.contentId as fromId, a.title as fromTitle, a.contentSummary as fromSummary,
				       b.contentId as toId,   b.title as toTitle,   b.contentSummary as toSummary,
				       r.linkId as linkId, type(r) as relationType, r.weight as weight,
				       r.createdAt as createdAt, r.updatedAt as updatedAt
				ORDER BY r.createdAt DESC
			`
		default:
			return nil, fmt.Errorf("invalid direction")
		}
		if filter.RelationType != "" {
			query += " AND type(r) = $relationType"
			params["relationType"] = string(filter.RelationType)
		}
		if filter.Limit > 0 {
			query += " LIMIT $limit"
			params["limit"] = filter.Limit
		}
		if filter.Offset > 0 {
			query += " SKIP $offset"
			params["offset"] = filter.Offset
		}
		records, runErr := tx.Run(ctx, query, params)
		if runErr != nil {
			return nil, runErr
		}
		var out []*domain.EnrichedKnowledgeLink
		for records.Next(ctx) {
			rec := records.Record()
			fromID, _ := uuid.Parse(rec.Values[0].(string))
			fromTitle, _ := rec.Values[1].(string)
			var fromSummary *string
			if v := rec.Values[2]; v != nil {
				if s, ok := v.(string); ok {
					fromSummary = &s
				}
			}
			toID, _ := uuid.Parse(rec.Values[3].(string))
			toTitle, _ := rec.Values[4].(string)
			var toSummary *string
			if v := rec.Values[5]; v != nil {
				if s, ok := v.(string); ok {
					toSummary = &s
				}
			}
			linkID, _ := uuid.Parse(rec.Values[6].(string))
			rel := domain.RelationType(rec.Values[7].(string))
			weight := rec.Values[8].(float64)
			createdAt := time.Unix(rec.Values[9].(int64), 0)
			updatedAt := time.Unix(rec.Values[10].(int64), 0)
			out = append(out, &domain.EnrichedKnowledgeLink{
				ID:           linkID,
				From:         domain.ContentPreview{ID: fromID, Title: fromTitle, ContentSummary: fromSummary},
				To:           domain.ContentPreview{ID: toID, Title: toTitle, ContentSummary: toSummary},
				RelationType: rel,
				Weight:       &weight,
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			})
		}
		return out, records.Err()
	})
	if err != nil {
		return nil, err
	}
	return result.([]*domain.EnrichedKnowledgeLink), nil
}

func (r *Neo4jRepository) UpdateKnowledgeLink(ctx context.Context, link *domain.KnowledgeLink) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Change relationship type by recreating the relationship with the new label
		valid, relType := normalizeAndValidateRelationship(string(link.RelationType))
		if !valid {
			return nil, fmt.Errorf("invalid relationship type: %s", link.RelationType)
		}

		query := fmt.Sprintf(`
			MATCH (a:Content)-[r {linkId: $linkId}]->(b:Content)
			WITH a, b, r, r.createdAt AS createdAt
			DELETE r
			CREATE (a)-[newRel:%s {
				linkId: $linkId,
				weight: $weight,
				createdAt: createdAt,
				updatedAt: $updatedAt
			}]->(b)
			RETURN newRel.linkId as linkId
		`, relType)

		var weight float64
		if link.Weight != nil {
			weight = *link.Weight
		} else {
			weight = 1.0
		}

		result, runErr := tx.Run(ctx, query, map[string]any{
			"linkId":    link.ID.String(),
			"weight":    weight,
			"updatedAt": link.UpdatedAt.Unix(),
		})

		if runErr != nil {
			return nil, runErr
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}
		return nil, fmt.Errorf("link not found")
	})
	return err
}

func (r *Neo4jRepository) DeleteKnowledgeLink(ctx context.Context, id uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, runErr := tx.Run(ctx,
			"MATCH ()-[r {linkId: $linkId}]->() DELETE r",
			map[string]any{"linkId": id.String()},
		)
		return nil, runErr
	})
	return err
}

// Convenience Methods
func (r *Neo4jRepository) CreateLink(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID) error {
	weight := 1.0
	link := &domain.KnowledgeLink{
		ID:            uuid.New(),
		FromContentID: fromContentID,
		ToContentID:   toContentID,
		RelationType:  domain.RelationTypeReferences,
		Weight:        &weight,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	return r.CreateKnowledgeLink(ctx, link)
}

func (r *Neo4jRepository) CreateLinkWithType(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID, relationshipType string) error {
	valid, rel := normalizeAndValidateRelationship(relationshipType)
	if !valid {
		return fmt.Errorf("invalid relationship type: %s", relationshipType)
	}

	weight := 1.0
	link := &domain.KnowledgeLink{
		ID:            uuid.New(),
		FromContentID: fromContentID,
		ToContentID:   toContentID,
		RelationType:  domain.RelationType(rel),
		Weight:        &weight,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	return r.CreateKnowledgeLink(ctx, link)
}

func (r *Neo4jRepository) GetBacklinks(ctx context.Context, contentID uuid.UUID) ([]uuid.UUID, error) {
	filter := domain.LinkFilter{
		ContentID: contentID,
		Direction: domain.LinkDirectionInbound,
		Limit:     100, // Reasonable limit for backlinks
	}

	links, err := r.ListKnowledgeLinks(ctx, filter)
	if err != nil {
		return nil, err
	}

	var backlinkIDs []uuid.UUID
	for _, link := range links {
		backlinkIDs = append(backlinkIDs, link.From.ID)
	}

	return backlinkIDs, nil
}

// Statistics
func (r *Neo4jRepository) CountLinksByContent(ctx context.Context, contentID uuid.UUID) (int64, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (a:Content)-[r]-(b:Content)
			WHERE a.contentId = $contentId OR b.contentId = $contentId
			RETURN count(r) as count
		`

		records, runErr := tx.Run(ctx, query, map[string]any{"contentId": contentID.String()})
		if runErr != nil {
			return nil, runErr
		}

		if records.Next(ctx) {
			return records.Record().Values[0].(int64), nil
		}
		return int64(0), nil
	})

	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (r *Neo4jRepository) CountLinksBySpace(ctx context.Context, spaceID uuid.UUID) (int64, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Count links between content nodes that belong to the same space
		query := `
			MATCH (a:Content)-[r]->(b:Content)
			WHERE a.spaceId = $spaceId AND b.spaceId = $spaceId
			RETURN count(r) as count
		`

		records, runErr := tx.Run(ctx, query, map[string]any{"spaceId": spaceID.String()})
		if runErr != nil {
			return nil, runErr
		}

		if records.Next(ctx) {
			return records.Record().Values[0].(int64), nil
		}
		return int64(0), nil
	})

	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (r *Neo4jRepository) ListKnowledgeLinksBySpace(ctx context.Context, spaceID uuid.UUID, relationType domain.RelationType, limit, offset int) ([]*domain.KnowledgeLink, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (a:Content)-[r]->(b:Content)
			WHERE a.spaceId = $spaceId AND b.spaceId = $spaceId
			RETURN a.contentId as fromId, b.contentId as toId, r.linkId as linkId,
			       type(r) as relationType, r.weight as weight,
			       r.createdAt as createdAt, r.updatedAt as updatedAt
			ORDER BY r.createdAt DESC
		`
		params := map[string]any{
			"spaceId": spaceID.String(),
		}
		if relationType != "" {
			query = strings.Replace(query, "RETURN", "AND type(r) = $relationType RETURN", 1)
			params["relationType"] = string(relationType)
		}
		if limit > 0 {
			query += " LIMIT $limit"
			params["limit"] = limit
		}
		if offset > 0 {
			query += " SKIP $offset"
			params["offset"] = offset
		}

		records, runErr := tx.Run(ctx, query, params)
		if runErr != nil {
			return nil, runErr
		}

		var links []*domain.KnowledgeLink
		for records.Next(ctx) {
			record := records.Record()
			fromID, _ := uuid.Parse(record.Values[0].(string))
			toID, _ := uuid.Parse(record.Values[1].(string))
			linkID, _ := uuid.Parse(record.Values[2].(string))
			rel := domain.RelationType(record.Values[3].(string))
			weight := record.Values[4].(float64)
			createdAt := time.Unix(record.Values[5].(int64), 0)
			updatedAt := time.Unix(record.Values[6].(int64), 0)

			links = append(links, &domain.KnowledgeLink{
				ID:            linkID,
				FromContentID: fromID,
				ToContentID:   toID,
				RelationType:  rel,
				Weight:        &weight,
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			})
		}
		return links, records.Err()
	})

	if err != nil {
		return nil, err
	}
	return result.([]*domain.KnowledgeLink), nil
}
