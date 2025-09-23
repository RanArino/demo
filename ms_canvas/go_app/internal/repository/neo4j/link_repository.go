package neo4j

import (
	"context"
	"time"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"

	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LinkRepo struct {
	driver *Driver
}

func NewLinkRepo(driver *Driver) *LinkRepo { return &LinkRepo{driver: driver} }

// CreateHierarchicalLinks creates :HIERARCHICAL_PARENT links from HierarchicalLink objects.
func (r *LinkRepo) CreateHierarchicalLinks(ctx context.Context, links []*canvasv1.HierarchicalLink) error {
	if len(links) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		items := make([]map[string]any, 0, len(links))
		for _, link := range links {
			createdAt := link.Base.CreatedAt.AsTime()
			if createdAt.IsZero() {
				createdAt = time.Now().UTC()
			}
			updatedAt := link.Base.UpdatedAt.AsTime()
			if updatedAt.IsZero() {
				updatedAt = createdAt
			}
			items = append(items, map[string]any{
				"src":                  link.Base.SourceId,
				"dst":                  link.Base.TargetId,
				"connection_type":      link.ConnectionType,
				"hierarchy_depth":      link.HierarchyDepth,
				"exploration_metadata": link.Base.ExplorationMetadata,
				"style_metadata":       link.Base.StyleMetadata,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
				"deleted_at":           link.Base.DeletedAt.AsTime(),
			})
		}
		params := map[string]any{
			"items": items,
		}
		_, err := tx.Run(ctx, `
            UNWIND $items AS item
            MATCH (s {id: item.src})
            MATCH (t {id: item.dst})
            MERGE (s)-[r:HIERARCHICAL_PARENT]->(t)
            SET r.connection_type = item.connection_type,
                r.hierarchy_depth = item.hierarchy_depth,
                r.exploration_metadata = item.exploration_metadata,
                r.style_metadata = item.style_metadata,
                r.created_at = datetime(item.created_at),
                r.updated_at = datetime(item.updated_at),
                r.deleted_at = CASE WHEN item.deleted_at IS NULL THEN NULL ELSE datetime(item.deleted_at) END
        `, params)
		return nil, err
	})
	return err
}

// CreateSemanticLinks creates :SEMANTIC_LINK edges with score.
func (r *LinkRepo) CreateSemanticLinks(ctx context.Context, links []*canvasv1.SemanticLink) error {
	if len(links) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"items": make([]map[string]any, 0, len(links))}
		for _, l := range links {
			createdAt := l.Base.CreatedAt.AsTime()
			if createdAt.IsZero() {
				createdAt = time.Now().UTC()
			}
			updatedAt := l.Base.UpdatedAt.AsTime()
			if updatedAt.IsZero() {
				updatedAt = createdAt
			}
			params["items"] = append(params["items"].([]map[string]any), map[string]any{
				"src":                  l.Base.SourceId,
				"dst":                  l.Base.TargetId,
				"connection_type":      l.ConnectionType,
				"strength_score":       l.StrengthScore,
				"similarity_score":     l.SimilarityScore,
				"abstraction_bridge":   l.AbstractionBridge,
				"hierarchical_bridge":  l.HierarchicalBridge,
				"exploration_metadata": l.Base.ExplorationMetadata,
				"semantic_tags":        l.SemanticTags,
				"style_metadata":       l.Base.StyleMetadata,
				"description":          l.Description,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
				"deleted_at":           l.Base.DeletedAt.AsTime(),
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

// CreateStructuralLinks creates :STRUCTURAL_LINK edges.
func (r *LinkRepo) CreateStructuralLinks(ctx context.Context, links []*canvasv1.StructuralLink) error {
	if len(links) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"items": make([]map[string]any, 0, len(links))}
		for _, l := range links {
			createdAt := l.Base.CreatedAt.AsTime()
			if createdAt.IsZero() {
				createdAt = time.Now().UTC()
			}
			updatedAt := l.Base.UpdatedAt.AsTime()
			if updatedAt.IsZero() {
				updatedAt = createdAt
			}
			params["items"] = append(params["items"].([]map[string]any), map[string]any{
				"src":                  l.Base.SourceId,
				"dst":                  l.Base.TargetId,
				"connection_type":      l.ConnectionType,
				"confidence_score":     l.ConfidenceScore,
				"description":          l.Description,
				"exploration_metadata": l.Base.ExplorationMetadata,
				"style_metadata":       l.Base.StyleMetadata,
				"created_by":           l.CreatedBy,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
				"deleted_at":           l.Base.DeletedAt.AsTime(),
			})
		}
		_, err := tx.Run(ctx, `
            UNWIND $items AS item
            MATCH (s {id: item.src})
            MATCH (t {id: item.dst})
            MERGE (s)-[r:STRUCTURAL_LINK]->(t)
            SET r.connection_type = item.connection_type,
                r.confidence_score = item.confidence_score,
                r.description = item.description,
                r.exploration_metadata = item.exploration_metadata,
                r.style_metadata = item.style_metadata,
                r.created_by = item.created_by,
                r.created_at = datetime(item.created_at),
                r.updated_at = datetime(item.updated_at),
                r.deleted_at = CASE WHEN item.deleted_at IS NULL THEN NULL ELSE datetime(item.deleted_at) END
        `, params)
		return nil, err
	})
	return err
}

// GetLinks returns links by IDs with optional filtering.
func (r *LinkRepo) GetLinks(ctx context.Context, ids []string, filter *canvasv1.BaseLinkFilter) ([]*canvasv1.Link, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	recAny, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"ids": ids}
		if filter != nil {
			params["source_id"] = filter.SourceId
			params["target_id"] = filter.TargetId
			params["connection_type"] = filter.ConnectionType
			params["exploration_metadata"] = filter.ExplorationMetadata
			params["style_metadata"] = filter.StyleMetadata
		}
		result, err := tx.Run(ctx, `
			// Query hierarchical links
			MATCH (s)-[r:HIERARCHICAL_PARENT]->(t)
			WHERE r.id IN $ids
			RETURN 'hierarchical' AS link_type,
				   r.id AS id,
				   s.id AS source_id,
				   t.id AS target_id,
				   r.connection_type AS connection_type,
				   r.hierarchy_depth AS hierarchy_depth,
				   r.exploration_metadata AS exploration_metadata,
				   r.style_metadata AS style_metadata,
				   r.created_at AS created_at,
				   r.updated_at AS updated_at,
				   r.deleted_at AS deleted_at
			UNION ALL
			// Query semantic links
			MATCH (s)-[r:SEMANTIC_LINK]->(t)
			WHERE r.id IN $ids
			RETURN 'semantic' AS link_type,
				   r.id AS id,
				   s.id AS source_id,
				   t.id AS target_id,
				   r.connection_type AS connection_type,
				   r.strength_score AS strength_score,
				   r.similarity_score AS similarity_score,
				   r.abstraction_bridge AS abstraction_bridge,
				   r.hierarchical_bridge AS hierarchical_bridge,
				   r.semantic_tags AS semantic_tags,
				   r.description AS description,
				   r.exploration_metadata AS exploration_metadata,
				   r.style_metadata AS style_metadata,
				   r.created_at AS created_at,
				   r.updated_at AS updated_at,
				   r.deleted_at AS deleted_at
			UNION ALL
			// Query structural links
			MATCH (s)-[r:STRUCTURAL_LINK]->(t)
			WHERE r.id IN $ids
			RETURN 'structural' AS link_type,
				   r.id AS id,
				   s.id AS source_id,
				   t.id AS target_id,
				   r.connection_type AS connection_type,
				   r.confidence_score AS confidence_score,
				   r.description AS description,
				   r.created_by AS created_by,
				   r.exploration_metadata AS exploration_metadata,
				   r.style_metadata AS style_metadata,
				   r.created_at AS created_at,
				   r.updated_at AS updated_at,
				   r.deleted_at AS deleted_at
		`, params)
		if err != nil {
			return nil, err
		}
		rows := make([]map[string]any, 0)
		for result.Next(ctx) {
			rec := result.Record()
			linkType, _ := rec.Get("link_type")
			id, _ := rec.Get("id")
			sourceId, _ := rec.Get("source_id")
			targetId, _ := rec.Get("target_id")
			connType, _ := rec.Get("connection_type")
			explorationMetadata, _ := rec.Get("exploration_metadata")
			styleMetadata, _ := rec.Get("style_metadata")
			createdAt, _ := rec.Get("created_at")
			updatedAt, _ := rec.Get("updated_at")
			deletedAt, _ := rec.Get("deleted_at")

			row := map[string]any{
				"link_type":            linkType,
				"id":                   id,
				"source_id":            sourceId,
				"target_id":            targetId,
				"connection_type":      connType,
				"exploration_metadata": explorationMetadata,
				"style_metadata":       styleMetadata,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
				"deleted_at":           deletedAt,
			}

			// Add type-specific fields
			switch linkType {
			case "hierarchical":
				if depth, ok := rec.Get("hierarchy_depth"); ok {
					row["hierarchy_depth"] = depth
				}
			case "semantic":
				if strength, ok := rec.Get("strength_score"); ok {
					row["strength_score"] = strength
				}
				if similarity, ok := rec.Get("similarity_score"); ok {
					row["similarity_score"] = similarity
				}
				if abstraction, ok := rec.Get("abstraction_bridge"); ok {
					row["abstraction_bridge"] = abstraction
				}
				if hierarchical, ok := rec.Get("hierarchical_bridge"); ok {
					row["hierarchical_bridge"] = hierarchical
				}
				if tags, ok := rec.Get("semantic_tags"); ok {
					row["semantic_tags"] = tags
				}
				if desc, ok := rec.Get("description"); ok {
					row["description"] = desc
				}
			case "structural":
				if confidence, ok := rec.Get("confidence_score"); ok {
					row["confidence_score"] = confidence
				}
				if desc, ok := rec.Get("description"); ok {
					row["description"] = desc
				}
				if createdBy, ok := rec.Get("created_by"); ok {
					row["created_by"] = createdBy
				}
			}
			rows = append(rows, row)
		}
		return rows, result.Err()
	})
	if err != nil {
		return nil, err
	}
	if recAny == nil {
		return nil, nil
	}

	rows := recAny.([]map[string]any)
	links := make([]*canvasv1.Link, 0, len(rows))

	for _, row := range rows {
		linkType, _ := row["link_type"].(string)
		id, _ := row["id"].(string)
		sourceId, _ := row["source_id"].(string)
		targetId, _ := row["target_id"].(string)
		explorationMetadata, _ := row["exploration_metadata"].(map[string]any)
		styleMetadata, _ := row["style_metadata"].(map[string]any)
		createdAt, _ := row["created_at"].(time.Time)
		updatedAt, _ := row["updated_at"].(time.Time)
		deletedAt, _ := row["deleted_at"].(time.Time)

		// Convert metadata to structpb.Struct
		var explorationMetadataStruct *structpb.Struct
		if explorationMetadata != nil {
			if s, err := structpb.NewStruct(explorationMetadata); err == nil {
				explorationMetadataStruct = s
			}
		}
		var styleMetadataStruct *structpb.Struct
		if styleMetadata != nil {
			if s, err := structpb.NewStruct(styleMetadata); err == nil {
				styleMetadataStruct = s
			}
		}

		var link *canvasv1.Link
		switch linkType {
		case "hierarchical":
			hierarchyDepth, _ := row["hierarchy_depth"].(int32)
			link = &canvasv1.Link{
				Link: &canvasv1.Link_Hierarchical{
					Hierarchical: &canvasv1.HierarchicalLink{
						Base: &canvasv1.BaseLink{
							Id:                  id,
							SourceId:            sourceId,
							TargetId:            targetId,
							ExplorationMetadata: explorationMetadataStruct,
							StyleMetadata:       styleMetadataStruct,
							CreatedAt:           timestamppb.New(createdAt),
							UpdatedAt:           timestamppb.New(updatedAt),
							DeletedAt:           timestamppb.New(deletedAt),
						},
						HierarchyDepth: hierarchyDepth,
					},
				},
			}
		case "semantic":
			strengthScore, _ := row["strength_score"].(float64)
			similarityScore, _ := row["similarity_score"].(float64)
			abstractionBridge, _ := row["abstraction_bridge"].(bool)
			hierarchicalBridge, _ := row["hierarchical_bridge"].(bool)
			semanticTags, _ := row["semantic_tags"].([]any)
			description, _ := row["description"].(string)

			tagStrings := make([]string, 0)
			for _, t := range semanticTags {
				if s, ok := t.(string); ok {
					tagStrings = append(tagStrings, s)
				}
			}
			var descPtr *string
			if description != "" {
				descPtr = &description
			}

			link = &canvasv1.Link{
				Link: &canvasv1.Link_Semantic{
					Semantic: &canvasv1.SemanticLink{
						Base: &canvasv1.BaseLink{
							Id:                  id,
							SourceId:            sourceId,
							TargetId:            targetId,
							ExplorationMetadata: explorationMetadataStruct,
							StyleMetadata:       styleMetadataStruct,
							CreatedAt:           timestamppb.New(createdAt),
							UpdatedAt:           timestamppb.New(updatedAt),
							DeletedAt:           timestamppb.New(deletedAt),
						},
						StrengthScore:      strengthScore,
						SimilarityScore:    similarityScore,
						AbstractionBridge:  abstractionBridge,
						HierarchicalBridge: hierarchicalBridge,
						SemanticTags:       tagStrings,
						Description:        descPtr,
					},
				},
			}
		case "structural":
			confidenceScore, _ := row["confidence_score"].(float64)
			description, _ := row["description"].(string)
			createdBy, _ := row["created_by"].(string)

			var descPtr *string
			if description != "" {
				descPtr = &description
			}

			link = &canvasv1.Link{
				Link: &canvasv1.Link_Structural{
					Structural: &canvasv1.StructuralLink{
						Base: &canvasv1.BaseLink{
							Id:                  id,
							SourceId:            sourceId,
							TargetId:            targetId,
							ExplorationMetadata: explorationMetadataStruct,
							StyleMetadata:       styleMetadataStruct,
							CreatedAt:           timestamppb.New(createdAt),
							UpdatedAt:           timestamppb.New(updatedAt),
							DeletedAt:           timestamppb.New(deletedAt),
						},
						ConfidenceScore: confidenceScore,
						Description:     descPtr,
						CreatedBy:       createdBy,
					},
				},
			}
		}

		if link != nil {
			links = append(links, link)
		}
	}

	return links, nil
}

// GetLinksByNodes returns links by node IDs with direction and filtering.
func (r *LinkRepo) GetLinksByNodes(ctx context.Context, nodeIDs []string, direction canvasv1.Direction, filter *canvasv1.BaseLinkFilter) ([]*canvasv1.Link, error) {
	if len(nodeIDs) == 0 {
		return nil, nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	recAny, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"ids": nodeIDs}
		if filter != nil {
			params["source_id"] = filter.SourceId
			params["target_id"] = filter.TargetId
			params["connection_type"] = filter.ConnectionType
			params["exploration_metadata"] = filter.ExplorationMetadata
			params["style_metadata"] = filter.StyleMetadata
		}
		result, err := tx.Run(ctx, `
			// Query hierarchical links
			MATCH (s)-[r:HIERARCHICAL_PARENT]->(t)
			WHERE s.id IN $ids OR t.id IN $ids
			RETURN 'hierarchical' AS link_type,
				   r.id AS id,
				   s.id AS source_id,
				   t.id AS target_id,
				   r.connection_type AS connection_type,
				   r.hierarchy_depth AS hierarchy_depth,
				   r.exploration_metadata AS exploration_metadata,
				   r.style_metadata AS style_metadata,
				   r.created_at AS created_at,
				   r.updated_at AS updated_at,
				   r.deleted_at AS deleted_at
			UNION ALL
			// Query semantic links
			MATCH (s)-[r:SEMANTIC_LINK]->(t)
			WHERE s.id IN $ids OR t.id IN $ids
			RETURN 'semantic' AS link_type,
				   r.id AS id,
				   s.id AS source_id,
				   t.id AS target_id,
				   r.connection_type AS connection_type,
				   r.strength_score AS strength_score,
				   r.similarity_score AS similarity_score,
				   r.abstraction_bridge AS abstraction_bridge,
				   r.hierarchical_bridge AS hierarchical_bridge,
				   r.semantic_tags AS semantic_tags,
				   r.description AS description,
				   r.exploration_metadata AS exploration_metadata,
				   r.style_metadata AS style_metadata,
				   r.created_at AS created_at,
				   r.updated_at AS updated_at,
				   r.deleted_at AS deleted_at
			UNION ALL
			// Query structural links
			MATCH (s)-[r:STRUCTURAL_LINK]->(t)
			WHERE s.id IN $ids OR t.id IN $ids
			RETURN 'structural' AS link_type,
				   r.id AS id,
				   s.id AS source_id,
				   t.id AS target_id,
				   r.connection_type AS connection_type,
				   r.confidence_score AS confidence_score,
				   r.description AS description,
				   r.created_by AS created_by,
				   r.exploration_metadata AS exploration_metadata,
				   r.style_metadata AS style_metadata,
				   r.created_at AS created_at,
				   r.updated_at AS updated_at,
				   r.deleted_at AS deleted_at
		`, params)
		if err != nil {
			return nil, err
		}
		rows := make([]map[string]any, 0)
		for result.Next(ctx) {
			rec := result.Record()
			linkType, _ := rec.Get("link_type")
			id, _ := rec.Get("id")
			sourceId, _ := rec.Get("source_id")
			targetId, _ := rec.Get("target_id")
			connType, _ := rec.Get("connection_type")
			explorationMetadata, _ := rec.Get("exploration_metadata")
			styleMetadata, _ := rec.Get("style_metadata")
			createdAt, _ := rec.Get("created_at")
			updatedAt, _ := rec.Get("updated_at")
			deletedAt, _ := rec.Get("deleted_at")

			row := map[string]any{
				"link_type":            linkType,
				"id":                   id,
				"source_id":            sourceId,
				"target_id":            targetId,
				"connection_type":      connType,
				"exploration_metadata": explorationMetadata,
				"style_metadata":       styleMetadata,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
				"deleted_at":           deletedAt,
			}

			// Add type-specific fields
			switch linkType {
			case "hierarchical":
				if depth, ok := rec.Get("hierarchy_depth"); ok {
					row["hierarchy_depth"] = depth
				}
			case "semantic":
				if strength, ok := rec.Get("strength_score"); ok {
					row["strength_score"] = strength
				}
				if similarity, ok := rec.Get("similarity_score"); ok {
					row["similarity_score"] = similarity
				}
				if abstraction, ok := rec.Get("abstraction_bridge"); ok {
					row["abstraction_bridge"] = abstraction
				}
				if hierarchical, ok := rec.Get("hierarchical_bridge"); ok {
					row["hierarchical_bridge"] = hierarchical
				}
				if tags, ok := rec.Get("semantic_tags"); ok {
					row["semantic_tags"] = tags
				}
				if desc, ok := rec.Get("description"); ok {
					row["description"] = desc
				}
			case "structural":
				if confidence, ok := rec.Get("confidence_score"); ok {
					row["confidence_score"] = confidence
				}
				if desc, ok := rec.Get("description"); ok {
					row["description"] = desc
				}
				if createdBy, ok := rec.Get("created_by"); ok {
					row["created_by"] = createdBy
				}
			}
			rows = append(rows, row)
		}
		return rows, result.Err()
	})
	if err != nil {
		return nil, err
	}
	if recAny == nil {
		return nil, nil
	}

	rows := recAny.([]map[string]any)
	links := make([]*canvasv1.Link, 0, len(rows))

	for _, row := range rows {
		linkType, _ := row["link_type"].(string)
		id, _ := row["id"].(string)
		sourceId, _ := row["source_id"].(string)
		targetId, _ := row["target_id"].(string)
		explorationMetadata, _ := row["exploration_metadata"].(map[string]any)
		styleMetadata, _ := row["style_metadata"].(map[string]any)
		createdAt, _ := row["created_at"].(time.Time)
		updatedAt, _ := row["updated_at"].(time.Time)
		deletedAt, _ := row["deleted_at"].(time.Time)

		// Convert metadata to structpb.Struct
		var explorationMetadataStruct *structpb.Struct
		if explorationMetadata != nil {
			if s, err := structpb.NewStruct(explorationMetadata); err == nil {
				explorationMetadataStruct = s
			}
		}
		var styleMetadataStruct *structpb.Struct
		if styleMetadata != nil {
			if s, err := structpb.NewStruct(styleMetadata); err == nil {
				styleMetadataStruct = s
			}
		}

		var link *canvasv1.Link
		switch linkType {
		case "hierarchical":
			hierarchyDepth, _ := row["hierarchy_depth"].(int32)
			link = &canvasv1.Link{
				Link: &canvasv1.Link_Hierarchical{
					Hierarchical: &canvasv1.HierarchicalLink{
						Base: &canvasv1.BaseLink{
							Id:                  id,
							SourceId:            sourceId,
							TargetId:            targetId,
							ExplorationMetadata: explorationMetadataStruct,
							StyleMetadata:       styleMetadataStruct,
							CreatedAt:           timestamppb.New(createdAt),
							UpdatedAt:           timestamppb.New(updatedAt),
							DeletedAt:           timestamppb.New(deletedAt),
						},
						HierarchyDepth: hierarchyDepth,
					},
				},
			}
		case "semantic":
			strengthScore, _ := row["strength_score"].(float64)
			similarityScore, _ := row["similarity_score"].(float64)
			abstractionBridge, _ := row["abstraction_bridge"].(bool)
			hierarchicalBridge, _ := row["hierarchical_bridge"].(bool)
			semanticTags, _ := row["semantic_tags"].([]any)
			description, _ := row["description"].(string)

			tagStrings := make([]string, 0)
			for _, t := range semanticTags {
				if s, ok := t.(string); ok {
					tagStrings = append(tagStrings, s)
				}
			}
			var descPtr *string
			if description != "" {
				descPtr = &description
			}

			link = &canvasv1.Link{
				Link: &canvasv1.Link_Semantic{
					Semantic: &canvasv1.SemanticLink{
						Base: &canvasv1.BaseLink{
							Id:                  id,
							SourceId:            sourceId,
							TargetId:            targetId,
							ExplorationMetadata: explorationMetadataStruct,
							StyleMetadata:       styleMetadataStruct,
							CreatedAt:           timestamppb.New(createdAt),
							UpdatedAt:           timestamppb.New(updatedAt),
							DeletedAt:           timestamppb.New(deletedAt),
						},
						StrengthScore:      strengthScore,
						SimilarityScore:    similarityScore,
						AbstractionBridge:  abstractionBridge,
						HierarchicalBridge: hierarchicalBridge,
						SemanticTags:       tagStrings,
						Description:        descPtr,
					},
				},
			}
		case "structural":
			confidenceScore, _ := row["confidence_score"].(float64)
			description, _ := row["description"].(string)
			createdBy, _ := row["created_by"].(string)

			var descPtr *string
			if description != "" {
				descPtr = &description
			}

			link = &canvasv1.Link{
				Link: &canvasv1.Link_Structural{
					Structural: &canvasv1.StructuralLink{
						Base: &canvasv1.BaseLink{
							Id:                  id,
							SourceId:            sourceId,
							TargetId:            targetId,
							ExplorationMetadata: explorationMetadataStruct,
							StyleMetadata:       styleMetadataStruct,
							CreatedAt:           timestamppb.New(createdAt),
							UpdatedAt:           timestamppb.New(updatedAt),
							DeletedAt:           timestamppb.New(deletedAt),
						},
						ConfidenceScore: confidenceScore,
						Description:     descPtr,
						CreatedBy:       createdBy,
					},
				},
			}
		}

		if link != nil {
			links = append(links, link)
		}
	}

	return links, nil
}

// UpdateSemanticLinks updates properties on existing semantic links.
func (r *LinkRepo) UpdateSemanticLinks(ctx context.Context, links []*canvasv1.SemanticLink) error {
	if len(links) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"items": make([]map[string]any, 0, len(links))}
		for _, l := range links {
			params["items"] = append(params["items"].([]map[string]any), map[string]any{
				"src":              l.Base.SourceId,
				"dst":              l.Base.TargetId,
				"connection_type":  l.ConnectionType,
				"strength_score":   l.StrengthScore,
				"similarity_score": l.SimilarityScore,
			})
		}
		_, err := tx.Run(ctx, `
            UNWIND $items AS item
            MATCH (s {id: item.src})-[r:SEMANTIC_LINK]->(t {id: item.dst})
            SET r.connection_type = item.connection_type,
                r.strength_score = item.strength_score,
                r.similarity_score = item.similarity_score
        `, params)
		return nil, err
	})
	return err
}

// UpdateStructuralLinks updates properties on existing structural links.
func (r *LinkRepo) UpdateStructuralLinks(ctx context.Context, links []*canvasv1.StructuralLink) error {
	if len(links) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"items": make([]map[string]any, 0, len(links))}
		for _, l := range links {
			params["items"] = append(params["items"].([]map[string]any), map[string]any{
				"src":              l.Base.SourceId,
				"dst":              l.Base.TargetId,
				"connection_type":  l.ConnectionType,
				"confidence_score": l.ConfidenceScore,
				"description":      l.Description,
				"created_by":       l.CreatedBy,
			})
		}
		_, err := tx.Run(ctx, `
            UNWIND $items AS item
            MATCH (s {id: item.src})-[r:STRUCTURAL_LINK]->(t {id: item.dst})
            SET r.connection_type = item.connection_type,
                r.confidence_score = item.confidence_score,
                r.description = item.description,
                r.created_by = item.created_by
        `, params)
		return nil, err
	})
	return err
}

// UpdateHierarchicalLinks updates properties on existing hierarchical links.
func (r *LinkRepo) UpdateHierarchicalLinks(ctx context.Context, links []*canvasv1.HierarchicalLink) error {
	if len(links) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"items": make([]map[string]any, 0, len(links))}
		for _, l := range links {
			params["items"] = append(params["items"].([]map[string]any), map[string]any{
				"src":             l.Base.SourceId,
				"dst":             l.Base.TargetId,
				"connection_type": l.ConnectionType,
				"hierarchy_depth": l.HierarchyDepth,
			})
		}
		_, err := tx.Run(ctx, `
            UNWIND $items AS item
            MATCH (s {id: item.src})-[r:HIERARCHICAL_PARENT]->(t {id: item.dst})
            SET r.connection_type = item.connection_type,
                r.hierarchy_depth = item.hierarchy_depth
        `, params)
		return nil, err
	})
	return err
}

// DeleteLinks deletes links by IDs.
func (r *LinkRepo) DeleteLinks(ctx context.Context, linkIDs []string) error {
	if len(linkIDs) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"ids": linkIDs}
		_, err := tx.Run(ctx, `
            MATCH ()-[r]-()
            WHERE r.id IN $ids
            DELETE r
        `, params)
		return nil, err
	})
	return err
}

// DeleteLinksForNodes deletes all links for the given nodes.
func (r *LinkRepo) DeleteLinksForNodes(ctx context.Context, nodeIDs []string) error {
	if len(nodeIDs) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		params := map[string]any{"ids": nodeIDs}
		_, err := tx.Run(ctx, `
            MATCH (n {id: $ids})-[r]-()
            DELETE r
        `, params)
		return nil, err
	})
	return err
}
