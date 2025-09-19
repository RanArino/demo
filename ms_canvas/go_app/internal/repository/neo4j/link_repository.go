package neo4j

import (
	"context"
	"time"

	"demo/ms_canvas/go_app/internal/domain"

	"github.com/google/uuid"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type LinkRepo struct {
	driver *Driver
}

func NewLinkRepo(driver *Driver) *LinkRepo { return &LinkRepo{driver: driver} }

// CreateHierarchicalLinks creates :HIERARCHICAL_PARENT links from ContentNode to given ChunkNode IDs.
func (r *LinkRepo) CreateHierarchicalLinks(contentSourceID uuid.UUID, chunkIDs []uuid.UUID, connectionType string, hierarchyDepth int, createdAt time.Time) error {
	if len(chunkIDs) == 0 {
		return nil
	}
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		items := make([]map[string]any, 0, len(chunkIDs))
		for _, id := range chunkIDs {
			items = append(items, map[string]any{"chunk_id": id.String()})
		}
		params := map[string]any{
			"content_source_id": contentSourceID.String(),
			"items":             items,
			"connection_type":   connectionType,
			"hierarchy_depth":   hierarchyDepth,
			"created_at":        createdAt,
		}
		_, err := tx.Run(ctx, `
            MATCH (c:ContentNode {content_source_id: $content_source_id})
            UNWIND $items AS item
            MATCH (n:ChunkNode {id: item.chunk_id})
            MERGE (c)-[r:HIERARCHICAL_PARENT]->(n)
            SET r.connection_type = $connection_type,
                r.hierarchy_depth = $hierarchy_depth,
                r.created_at = datetime($created_at)
        `, params)
		return nil, err
	})
	return err
}

// CreateSemanticLinks creates :SEMANTIC_LINK edges with score.
func (r *LinkRepo) CreateSemanticLinks(links []domain.SemanticLink) error {
	if len(links) == 0 {
		return nil
	}
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"items": make([]map[string]any, 0, len(links))}
		for _, l := range links {
			createdAt := l.CreatedAt
			if createdAt.IsZero() {
				createdAt = time.Now().UTC()
			}
			updatedAt := l.UpdatedAt
			if updatedAt.IsZero() {
				updatedAt = createdAt
			}
			params["items"] = append(params["items"].([]map[string]any), map[string]any{
				"src":                  l.SourceID.String(),
				"dst":                  l.TargetID.String(),
				"connection_type":      l.ConnectionType,
				"strength_score":       l.StrengthScore,
				"similarity_score":     l.SimilarityScore,
				"abstraction_bridge":   l.AbstractionBridge,
				"hierarchical_bridge":  l.HierarchicalBridge,
				"exploration_metadata": l.ExplorationMetadata,
				"semantic_tags":        l.SemanticTags,
				"style_metadata":       l.StyleMetadata,
				"description":          l.Description,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
				"deleted_at":           l.DeletedAt,
			})
		}
		_, err := tx.Run(ctx, `
            UNWIND $items AS item
            MATCH (s {id: item.src})
            MATCH (t {id: item.dst})
            MERGE (s)-[r:SEMANTIC_LINK]->(t)
            SET r.connection_type = item.connection_type,
                r.strength_score = item.strength_score,
                r.similarity_score = item.similarity_score,
                r.abstraction_bridge = item.abstraction_bridge,
                r.hierarchical_bridge = item.hierarchical_bridge,
                r.exploration_metadata = item.exploration_metadata,
                r.semantic_tags = item.semantic_tags,
                r.style_metadata = item.style_metadata,
                r.description = item.description,
                r.created_at = datetime(item.created_at),
                r.updated_at = datetime(item.updated_at),
                r.deleted_at = CASE WHEN item.deleted_at IS NULL THEN NULL ELSE datetime(item.deleted_at) END
        `, params)
		return nil, err
	})
	return err
}

// CreateStructuralLink creates an explicit STRUCTURAL_LINK edge.
func (r *LinkRepo) CreateStructuralLink(link domain.StructuralLink) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		createdAt := link.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		updatedAt := link.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}
		params := map[string]any{
			"src":                  link.SourceID.String(),
			"dst":                  link.TargetID.String(),
			"connection_type":      link.ConnectionType,
			"confidence_score":     link.ConfidenceScore,
			"description":          link.Description,
			"exploration_metadata": link.ExplorationMetadata,
			"style_metadata":       link.StyleMetadata,
			"created_by":           link.CreatedBy,
			"created_at":           createdAt,
			"updated_at":           updatedAt,
			"deleted_at":           link.DeletedAt,
		}
		_, err := tx.Run(ctx, `
            MATCH (s {id: $src})
            MATCH (t {id: $dst})
            MERGE (s)-[r:STRUCTURAL_LINK]->(t)
            SET r.connection_type = $connection_type,
                r.confidence_score = $confidence_score,
                r.description = $description,
                r.exploration_metadata = $exploration_metadata,
                r.style_metadata = $style_metadata,
                r.created_by = $created_by,
                r.created_at = datetime($created_at),
                r.updated_at = datetime($updated_at),
                r.deleted_at = CASE WHEN $deleted_at IS NULL THEN NULL ELSE datetime($deleted_at) END
        `, params)
		return nil, err
	})
	return err
}

// GetSemanticLink returns the semantic link between two nodes if it exists.
func (r *LinkRepo) GetSemanticLink(sourceID, targetID uuid.UUID) (*domain.SemanticLink, error) {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	recAny, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"src": sourceID.String(), "dst": targetID.String()}
		result, err := tx.Run(ctx, `
            MATCH (s {id: $src})-[r:SEMANTIC_LINK]->(t {id: $dst})
            RETURN r.connection_type AS connection_type,
                   r.strength_score AS strength_score,
                   r.similarity_score AS similarity_score,
                   r.abstraction_bridge AS abstraction_bridge,
                   r.hierarchical_bridge AS hierarchical_bridge,
                   r.exploration_metadata AS exploration_metadata,
                   r.semantic_tags AS semantic_tags,
                   r.style_metadata AS style_metadata,
                   r.description AS description,
                   r.created_at AS created_at,
                   r.updated_at AS updated_at,
                   r.deleted_at AS deleted_at
        `, params)
		if err != nil {
			return nil, err
		}
		if result.Next(ctx) {
			rec := result.Record()
			conn, _ := rec.Get("connection_type")
			strength, _ := rec.Get("strength_score")
			similarity, _ := rec.Get("similarity_score")
			abBridge, _ := rec.Get("abstraction_bridge")
			hBridge, _ := rec.Get("hierarchical_bridge")
			expl, _ := rec.Get("exploration_metadata")
			tags, _ := rec.Get("semantic_tags")
			style, _ := rec.Get("style_metadata")
			desc, _ := rec.Get("description")
			createdAt, _ := rec.Get("created_at")
			updatedAt, _ := rec.Get("updated_at")
			deletedAt, _ := rec.Get("deleted_at")
			return map[string]any{
				"connection_type":      conn,
				"strength_score":       strength,
				"similarity_score":     similarity,
				"abstraction_bridge":   abBridge,
				"hierarchical_bridge":  hBridge,
				"exploration_metadata": expl,
				"semantic_tags":        tags,
				"style_metadata":       style,
				"description":          desc,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
				"deleted_at":           deletedAt,
			}, result.Err()
		}
		return nil, result.Err()
	})
	if err != nil {
		return nil, err
	}
	if recAny == nil {
		return nil, nil
	}
	m := recAny.(map[string]any)
	conn, _ := m["connection_type"].(string)
	strength, _ := m["strength_score"].(float64)
	similarity, _ := m["similarity_score"].(float64)
	abBridge, _ := m["abstraction_bridge"].(bool)
	hBridge, _ := m["hierarchical_bridge"].(bool)
	expl, _ := m["exploration_metadata"].(map[string]any)
	tags, _ := m["semantic_tags"].([]any)
	style, _ := m["style_metadata"].(map[string]any)
	var tagStrings []string
	for _, t := range tags {
		if s, ok := t.(string); ok {
			tagStrings = append(tagStrings, s)
		}
	}
	var descPtr *string
	if v, ok := m["description"].(string); ok {
		descPtr = &v
	}
	var delPtr *time.Time
	if v, ok := m["deleted_at"].(time.Time); ok {
		delPtr = &v
	}
	createdAt, _ := m["created_at"].(time.Time)
	updatedAt, _ := m["updated_at"].(time.Time)
	link := domain.SemanticLink{
		SourceID:            sourceID,
		TargetID:            targetID,
		ConnectionType:      conn,
		StrengthScore:       strength,
		SimilarityScore:     similarity,
		AbstractionBridge:   abBridge,
		HierarchicalBridge:  hBridge,
		ExplorationMetadata: expl,
		SemanticTags:        tagStrings,
		StyleMetadata:       style,
		Description:         descPtr,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		DeletedAt:           delPtr,
	}
	return &link, nil
}

// ListSemanticLinksFrom lists outgoing semantic links from the given node.
func (r *LinkRepo) ListSemanticLinksFrom(nodeID uuid.UUID) ([]domain.SemanticLink, error) {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	recAny, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"id": nodeID.String()}
		result, err := tx.Run(ctx, `
            MATCH (s {id: $id})-[r:SEMANTIC_LINK]->(t)
            RETURN t.id AS target,
                   r.connection_type AS connection_type,
                   r.strength_score AS strength_score,
                   r.similarity_score AS similarity_score
        `, params)
		if err != nil {
			return nil, err
		}
		rows := make([]map[string]any, 0)
		for result.Next(ctx) {
			rec := result.Record()
			target, _ := rec.Get("target")
			conn, _ := rec.Get("connection_type")
			strength, _ := rec.Get("strength_score")
			similarity, _ := rec.Get("similarity_score")
			rows = append(rows, map[string]any{
				"target":           target,
				"connection_type":  conn,
				"strength_score":   strength,
				"similarity_score": similarity,
			})
		}
		return rows, result.Err()
	})
	if err != nil {
		return nil, err
	}
	rows := recAny.([]map[string]any)
	out := make([]domain.SemanticLink, 0, len(rows))
	for _, rmap := range rows {
		targetStr, _ := rmap["target"].(string)
		targetID, _ := uuid.Parse(targetStr)
		conn, _ := rmap["connection_type"].(string)
		strength, _ := rmap["strength_score"].(float64)
		similarity, _ := rmap["similarity_score"].(float64)
		out = append(out, domain.SemanticLink{SourceID: nodeID, TargetID: targetID, ConnectionType: conn, StrengthScore: strength, SimilarityScore: similarity})
	}
	return out, nil
}

// ListSemanticLinksTo lists incoming semantic links to the given node.
func (r *LinkRepo) ListSemanticLinksTo(nodeID uuid.UUID) ([]domain.SemanticLink, error) {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	recAny, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"id": nodeID.String()}
		result, err := tx.Run(ctx, `
            MATCH (s)-[r:SEMANTIC_LINK]->(t {id: $id})
            RETURN s.id AS source,
                   r.connection_type AS connection_type,
                   r.strength_score AS strength_score,
                   r.similarity_score AS similarity_score
        `, params)
		if err != nil {
			return nil, err
		}
		rows := make([]map[string]any, 0)
		for result.Next(ctx) {
			rec := result.Record()
			source, _ := rec.Get("source")
			conn, _ := rec.Get("connection_type")
			strength, _ := rec.Get("strength_score")
			similarity, _ := rec.Get("similarity_score")
			rows = append(rows, map[string]any{
				"source":           source,
				"connection_type":  conn,
				"strength_score":   strength,
				"similarity_score": similarity,
			})
		}
		return rows, result.Err()
	})
	if err != nil {
		return nil, err
	}
	rows := recAny.([]map[string]any)
	out := make([]domain.SemanticLink, 0, len(rows))
	for _, rmap := range rows {
		sourceStr, _ := rmap["source"].(string)
		sourceID, _ := uuid.Parse(sourceStr)
		conn, _ := rmap["connection_type"].(string)
		strength, _ := rmap["strength_score"].(float64)
		similarity, _ := rmap["similarity_score"].(float64)
		out = append(out, domain.SemanticLink{SourceID: sourceID, TargetID: nodeID, ConnectionType: conn, StrengthScore: strength, SimilarityScore: similarity})
	}
	return out, nil
}

// UpdateSemanticLink updates properties on an existing semantic link.
func (r *LinkRepo) UpdateSemanticLink(link domain.SemanticLink) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{
			"src":              link.SourceID.String(),
			"dst":              link.TargetID.String(),
			"connection_type":  link.ConnectionType,
			"strength_score":   link.StrengthScore,
			"similarity_score": link.SimilarityScore,
		}
		_, err := tx.Run(ctx, `
            MATCH (s {id: $src})-[r:SEMANTIC_LINK]->(t {id: $dst})
            SET r.connection_type = $connection_type,
                r.strength_score = $strength_score,
                r.similarity_score = $similarity_score
        `, params)
		return nil, err
	})
	return err
}

// DeleteSemanticLink deletes a single semantic link between two nodes.
func (r *LinkRepo) DeleteSemanticLink(sourceID, targetID uuid.UUID) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"src": sourceID.String(), "dst": targetID.String()}
		_, err := tx.Run(ctx, `
            MATCH (s {id: $src})-[r:SEMANTIC_LINK]->(t {id: $dst})
            DELETE r
        `, params)
		return nil, err
	})
	return err
}

// DeleteSemanticLinksForNode deletes all incoming and outgoing semantic links for a node.
func (r *LinkRepo) DeleteSemanticLinksForNode(nodeID uuid.UUID) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"id": nodeID.String()}
		_, err := tx.Run(ctx, `
            MATCH (n {id: $id})-[r:SEMANTIC_LINK]-()
            DELETE r
        `, params)
		return nil, err
	})
	return err
}

// Structural link queries

// GetStructuralLink returns the structural link between two nodes if it exists.
func (r *LinkRepo) GetStructuralLink(sourceID, targetID uuid.UUID) (*domain.StructuralLink, error) {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	recAny, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"src": sourceID.String(), "dst": targetID.String()}
		result, err := tx.Run(ctx, `
            MATCH (s {id: $src})-[r:STRUCTURAL_LINK]->(t {id: $dst})
            RETURN r.connection_type AS connection_type,
                   r.confidence_score AS confidence_score,
                   r.description AS description,
                   r.exploration_metadata AS exploration_metadata,
                   r.style_metadata AS style_metadata,
                   r.created_by AS created_by,
                   r.created_at AS created_at,
                   r.updated_at AS updated_at,
                   r.deleted_at AS deleted_at
        `, params)
		if err != nil {
			return nil, err
		}
		if result.Next(ctx) {
			rec := result.Record()
			conn, _ := rec.Get("connection_type")
			conf, _ := rec.Get("confidence_score")
			desc, _ := rec.Get("description")
			expl, _ := rec.Get("exploration_metadata")
			style, _ := rec.Get("style_metadata")
			createdBy, _ := rec.Get("created_by")
			createdAt, _ := rec.Get("created_at")
			updatedAt, _ := rec.Get("updated_at")
			deletedAt, _ := rec.Get("deleted_at")
			return map[string]any{
				"connection_type":      conn,
				"confidence_score":     conf,
				"description":          desc,
				"exploration_metadata": expl,
				"style_metadata":       style,
				"created_by":           createdBy,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
				"deleted_at":           deletedAt,
			}, result.Err()
		}
		return nil, result.Err()
	})
	if err != nil {
		return nil, err
	}
	if recAny == nil {
		return nil, nil
	}
	m := recAny.(map[string]any)
	conn, _ := m["connection_type"].(string)
	conf, _ := m["confidence_score"].(float64)
	var descPtr *string
	if v, ok := m["description"].(string); ok {
		descPtr = &v
	}
	expl, _ := m["exploration_metadata"].(map[string]any)
	style, _ := m["style_metadata"].(map[string]any)
	createdBy, _ := m["created_by"].(string)
	createdAt, _ := m["created_at"].(time.Time)
	updatedAt, _ := m["updated_at"].(time.Time)
	var delPtr *time.Time
	if v, ok := m["deleted_at"].(time.Time); ok {
		delPtr = &v
	}
	link := domain.StructuralLink{
		SourceID:            sourceID,
		TargetID:            targetID,
		ConnectionType:      conn,
		ConfidenceScore:     conf,
		Description:         descPtr,
		ExplorationMetadata: expl,
		StyleMetadata:       style,
		CreatedBy:           createdBy,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		DeletedAt:           delPtr,
	}
	return &link, nil
}

// ListStructuralLinksFrom lists outgoing structural links from the given node.
func (r *LinkRepo) ListStructuralLinksFrom(nodeID uuid.UUID) ([]domain.StructuralLink, error) {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	recAny, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"id": nodeID.String()}
		result, err := tx.Run(ctx, `
            MATCH (s {id: $id})-[r:STRUCTURAL_LINK]->(t)
            RETURN t.id AS target,
                   r.connection_type AS connection_type,
                   r.confidence_score AS confidence_score,
                   r.created_by AS created_by
        `, params)
		if err != nil {
			return nil, err
		}
		rows := make([]map[string]any, 0)
		for result.Next(ctx) {
			rec := result.Record()
			target, _ := rec.Get("target")
			conn, _ := rec.Get("connection_type")
			conf, _ := rec.Get("confidence_score")
			createdBy, _ := rec.Get("created_by")
			rows = append(rows, map[string]any{
				"target":           target,
				"connection_type":  conn,
				"confidence_score": conf,
				"created_by":       createdBy,
			})
		}
		return rows, result.Err()
	})
	if err != nil {
		return nil, err
	}
	rows := recAny.([]map[string]any)
	out := make([]domain.StructuralLink, 0, len(rows))
	for _, rmap := range rows {
		targetStr, _ := rmap["target"].(string)
		targetID, _ := uuid.Parse(targetStr)
		conn, _ := rmap["connection_type"].(string)
		conf, _ := rmap["confidence_score"].(float64)
		createdBy, _ := rmap["created_by"].(string)
		out = append(out, domain.StructuralLink{SourceID: nodeID, TargetID: targetID, ConnectionType: conn, ConfidenceScore: conf, CreatedBy: createdBy})
	}
	return out, nil
}

// ListStructuralLinksTo lists incoming structural links to the given node.
func (r *LinkRepo) ListStructuralLinksTo(nodeID uuid.UUID) ([]domain.StructuralLink, error) {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	recAny, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"id": nodeID.String()}
		result, err := tx.Run(ctx, `
            MATCH (s)-[r:STRUCTURAL_LINK]->(t {id: $id})
            RETURN s.id AS source,
                   r.connection_type AS connection_type,
                   r.confidence_score AS confidence_score,
                   r.created_by AS created_by
        `, params)
		if err != nil {
			return nil, err
		}
		rows := make([]map[string]any, 0)
		for result.Next(ctx) {
			rec := result.Record()
			source, _ := rec.Get("source")
			conn, _ := rec.Get("connection_type")
			conf, _ := rec.Get("confidence_score")
			createdBy, _ := rec.Get("created_by")
			rows = append(rows, map[string]any{
				"source":           source,
				"connection_type":  conn,
				"confidence_score": conf,
				"created_by":       createdBy,
			})
		}
		return rows, result.Err()
	})
	if err != nil {
		return nil, err
	}
	rows := recAny.([]map[string]any)
	out := make([]domain.StructuralLink, 0, len(rows))
	for _, rmap := range rows {
		sourceStr, _ := rmap["source"].(string)
		srcID, _ := uuid.Parse(sourceStr)
		conn, _ := rmap["connection_type"].(string)
		conf, _ := rmap["confidence_score"].(float64)
		createdBy, _ := rmap["created_by"].(string)
		out = append(out, domain.StructuralLink{SourceID: srcID, TargetID: nodeID, ConnectionType: conn, ConfidenceScore: conf, CreatedBy: createdBy})
	}
	return out, nil
}

// UpdateStructuralLink updates properties on an existing structural link.
func (r *LinkRepo) UpdateStructuralLink(link domain.StructuralLink) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{
			"src":              link.SourceID.String(),
			"dst":              link.TargetID.String(),
			"connection_type":  link.ConnectionType,
			"confidence_score": link.ConfidenceScore,
			"description":      link.Description,
			"created_by":       link.CreatedBy,
		}
		_, err := tx.Run(ctx, `
            MATCH (s {id: $src})-[r:STRUCTURAL_LINK]->(t {id: $dst})
            SET r.connection_type = $connection_type,
                r.confidence_score = $confidence_score,
                r.description = $description,
                r.created_by = $created_by
        `, params)
		return nil, err
	})
	return err
}

// DeleteStructuralLink deletes a single structural link between two nodes.
func (r *LinkRepo) DeleteStructuralLink(sourceID, targetID uuid.UUID) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"src": sourceID.String(), "dst": targetID.String()}
		_, err := tx.Run(ctx, `
            MATCH (s {id: $src})-[r:STRUCTURAL_LINK]->(t {id: $dst})
            DELETE r
        `, params)
		return nil, err
	})
	return err
}

// DeleteStructuralLinksForNode deletes all incoming and outgoing structural links for a node.
func (r *LinkRepo) DeleteStructuralLinksForNode(nodeID uuid.UUID) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"id": nodeID.String()}
		_, err := tx.Run(ctx, `
            MATCH (n {id: $id})-[r:STRUCTURAL_LINK]-()
            DELETE r
        `, params)
		return nil, err
	})
	return err
}

// DeleteHierarchicalLinksByContentSource deletes all hierarchical links for a given content source.
func (r *LinkRepo) DeleteHierarchicalLinksByContentSource(contentSourceID uuid.UUID) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"content_source_id": contentSourceID.String()}
		_, err := tx.Run(ctx, `
            MATCH (c:ContentNode {content_source_id: $content_source_id})-[r:HIERARCHICAL_PARENT]->()
            DELETE r
        `, params)
		return nil, err
	})
	return err
}

// DeleteHierarchicalLinks deletes hierarchical links from a specific content source to a subset of chunk nodes.
func (r *LinkRepo) DeleteHierarchicalLinks(contentSourceID uuid.UUID, chunkIDs []uuid.UUID) error {
	if len(chunkIDs) == 0 {
		return nil
	}
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		items := make([]map[string]any, 0, len(chunkIDs))
		for _, id := range chunkIDs {
			items = append(items, map[string]any{"chunk_id": id.String()})
		}
		params := map[string]any{
			"content_source_id": contentSourceID.String(),
			"items":             items,
		}
		_, err := tx.Run(ctx, `
            UNWIND $items AS item
            MATCH (c:ContentNode {content_source_id: $content_source_id})-[r:HIERARCHICAL_PARENT]->(n {id: item.chunk_id})
            DELETE r
        `, params)
		return nil, err
	})
	return err
}
