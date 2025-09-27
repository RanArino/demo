package neo4j

import (
	"context"
	"fmt"
	"strings"
	"time"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"

	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NodeRepo struct {
	driver *Driver
}

func NewNodeRepo(driver *Driver) *NodeRepo { return &NodeRepo{driver: driver} }

// GetNodes retrieves nodes by their IDs with optional filtering.
func (r *NodeRepo) GetNodes(ctx context.Context, ids []string, filter *canvasv1.NodeFilter) ([]*canvasv1.Node, error) {
	if len(ids) == 0 {
		return []*canvasv1.Node{}, nil
	}

	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)

	result, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"ids": ids,
		}

		// Build WHERE clause for filtering
		whereConditions := []string{"n.id IN $ids"}

		if filter != nil {
			if filter.SpaceId != nil {
				params["space_id"] = *filter.SpaceId
				whereConditions = append(whereConditions, "n.space_id = $space_id")
			}
			if filter.AbstractionLevelMin != nil {
				params["abstraction_level_min"] = *filter.AbstractionLevelMin
				whereConditions = append(whereConditions, "n.abstraction_level >= $abstraction_level_min")
			}
			if filter.AbstractionLevelMax != nil {
				params["abstraction_level_max"] = *filter.AbstractionLevelMax
				whereConditions = append(whereConditions, "n.abstraction_level <= $abstraction_level_max")
			}
			if filter.ContextType != nil {
				params["context_type"] = *filter.ContextType
				whereConditions = append(whereConditions, "n.context_type = $context_type")
			}
			if len(filter.Keywords) > 0 {
				params["keywords"] = filter.Keywords
				whereConditions = append(whereConditions, "ANY(keyword IN $keywords WHERE n.keywords CONTAINS keyword)")
			}
			if filter.SemanticDensityMin != nil {
				params["semantic_density_min"] = *filter.SemanticDensityMin
				whereConditions = append(whereConditions, "n.semantic_density >= $semantic_density_min")
			}
			if filter.SemanticDensityMax != nil {
				params["semantic_density_max"] = *filter.SemanticDensityMax
				whereConditions = append(whereConditions, "n.semantic_density <= $semantic_density_max")
			}
			if filter.IsPositionLocked != nil {
				params["is_position_locked"] = *filter.IsPositionLocked
				whereConditions = append(whereConditions, "n.is_position_locked = $is_position_locked")
			}
			if filter.Visibility != nil {
				params["visibility"] = *filter.Visibility
				whereConditions = append(whereConditions, "n.visibility = $visibility")
			}
			if filter.CreatedAfter != nil {
				params["created_after"] = *filter.CreatedAfter
				whereConditions = append(whereConditions, "n.created_at >= datetime($created_after)")
			}
			if filter.CreatedBefore != nil {
				params["created_before"] = *filter.CreatedBefore
				whereConditions = append(whereConditions, "n.created_at <= datetime($created_before)")
			}
			if filter.UpdatedAfter != nil {
				params["updated_after"] = *filter.UpdatedAfter
				whereConditions = append(whereConditions, "n.updated_at >= datetime($updated_after)")
			}
			if filter.UpdatedBefore != nil {
				params["updated_before"] = *filter.UpdatedBefore
				whereConditions = append(whereConditions, "n.updated_at <= datetime($updated_before)")
			}
		}

		whereClause := ""
		if len(whereConditions) > 0 {
			whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
		}

		query := `
			MATCH (n:Node)
			` + whereClause + `
			RETURN
				n.id as id,
				labels(n) as labels,
				n.content_source_id as content_source_id,
				n.space_id as space_id,
				n.abstraction_level as abstraction_level,
				n.context_type as context_type,
				n.embedding as embedding,
				n.keywords as keywords,
				n.chat_content as chat_content,
				n.display_content as display_content,
				n.semantic_density as semantic_density,
				n.created_at as created_at,
				n.updated_at as updated_at,
				n.deleted_at as deleted_at,
				// ContentNode specific
				n.title as title,
				n.media_type as media_type,
				n.source as source,
				n.token_count as token_count,
				// ChunkNode specific
				n.sequence_index as sequence_index,
				n.chunk_type as chunk_type,
				n.start_position as start_position,
				n.end_position as end_position,
				n.content as content,
				// ClusterNode specific
				n.cluster_scope as cluster_scope,
				n.member_count as member_count,
				n.coverage_score as coverage_score,
				// Position
				n.location as location
			ORDER BY n.id
		`

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var nodes []*canvasv1.Node
		for result.Next(ctx) {
			record := result.Record()

			labels, _ := record.Get("labels")
			labelList := labels.([]interface{})

			// Determine node type from labels
			var node *canvasv1.Node
			for _, label := range labelList {
				labelStr := label.(string)
				switch labelStr {
				case "ContentNode":
					node = r.buildContentNode(record)
				case "ChunkNode":
					node = r.buildChunkNode(record)
				case "ClusterNode":
					node = r.buildClusterNode(record)
				}
				if node != nil {
					break
				}
			}

			if node != nil {
				nodes = append(nodes, node)
			}
		}

		return nodes, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*canvasv1.Node), nil
}

// SearchNodes searches for nodes based on filter criteria and spatial bounds
func (r *NodeRepo) SearchNodes(ctx context.Context, filter *canvasv1.NodeFilter, spatialBBox *canvasv1.SpatialBoundingBox, limit int32) ([]*canvasv1.Node, error) {
	// Get nodes with filter
	nodes, err := r.GetNodes(ctx, []string{}, filter)
	if err != nil {
		return nil, err
	}

	// Apply spatial filtering if spatial bounding box is provided
	if spatialBBox != nil {
		nodes = r.applySpatialFilter(nodes, spatialBBox)
	}

	// Apply limit if specified
	if limit > 0 && int32(len(nodes)) > limit {
		nodes = nodes[:limit]
	}

	return nodes, nil
}

// applySpatialFilter filters nodes based on spatial bounding box
func (r *NodeRepo) applySpatialFilter(nodes []*canvasv1.Node, spatialBBox *canvasv1.SpatialBoundingBox) []*canvasv1.Node {
	if spatialBBox == nil {
		return nodes
	}

	var filtered []*canvasv1.Node
	for _, node := range nodes {
		if r.nodeWithinBounds(node, spatialBBox) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// nodeWithinBounds checks if a node's position is within the spatial bounding box
func (r *NodeRepo) nodeWithinBounds(node *canvasv1.Node, spatialBBox *canvasv1.SpatialBoundingBox) bool {
	var position *canvasv1.SpatialCoordinates
	var chunk *canvasv1.ChunkNode
	var content *canvasv1.ContentNode
	var cluster *canvasv1.ClusterNode

	// Extract position based on node type
	switch n := node.Node.(type) {
	case *canvasv1.Node_Chunk:
		chunk = n.Chunk
		position = chunk.Base.Position_3D
	case *canvasv1.Node_Content:
		content = n.Content
		position = content.Base.Position_3D
	case *canvasv1.Node_Cluster:
		cluster = n.Cluster
		position = cluster.Base.Position_3D
	default:
		return false // Unknown node type
	}

	if position == nil {
		return false // No position data
	}

	// Check if node position is within bounds
	return position.X >= spatialBBox.MinCoords.X &&
		position.X <= spatialBBox.MaxCoords.X &&
		position.Y >= spatialBBox.MinCoords.Y &&
		position.Y <= spatialBBox.MaxCoords.Y &&
		position.Z >= spatialBBox.MinCoords.Z &&
		position.Z <= spatialBBox.MaxCoords.Z
}

// CreateChunkNodes creates chunk nodes (idempotent on id/content_source_id via MERGE).
func (r *NodeRepo) CreateChunkNodes(ctx context.Context, chunks []*canvasv1.ChunkNode) error {
	if len(chunks) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"items": make([]map[string]interface{}, 0, len(chunks)),
		}
		now := time.Now().UTC().Format(time.RFC3339)
		for _, c := range chunks {
			id := c.Base.Id
			item := map[string]interface{}{
				"id":                id,
				"content_source_id": c.ContentSourceId,
				"sequence_index":    c.SequenceIndex,
				"start_position":    c.StartPosition,
				"end_position":      c.EndPosition,
				"content":           c.Content,
				"now":               now,
			}

			// Add location only if Position_3D is not nil
			if c.Base != nil && c.Base.Position_3D != nil {
				item["location"] = map[string]interface{}{
					"x": c.Base.Position_3D.X,
					"y": c.Base.Position_3D.Y,
					"z": c.Base.Position_3D.Z,
				}
			}

			params["items"] = append(params["items"].([]map[string]interface{}), item)
		}
		_, err := tx.Run(ctx, `
			UNWIND $items AS item
			MERGE (n:ChunkNode:Node {id: item.id})
			ON CREATE SET
				n.content_source_id = item.content_source_id,
				n.sequence_index = item.sequence_index,
				n.location = item.location,
				n.start_position = item.start_position,
				n.end_position = item.end_position,
				n.content = item.content,
				n.created_at = datetime(item.now),
				n.updated_at = datetime(item.now)
			ON MATCH SET
				n.content = item.content,
				n.location = item.location,
				n.sequence_index = item.sequence_index,
				n.updated_at = datetime(item.now)
		`, params)
		return nil, err
	})
	return err
}

// CreateContentNodes creates or updates ContentNodes in batch.
func (r *NodeRepo) CreateContentNodes(ctx context.Context, contents []*canvasv1.ContentNode) error {
	if len(contents) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"items": make([]map[string]interface{}, 0, len(contents)),
		}
		now := time.Now().UTC().Format(time.RFC3339)
		for _, content := range contents {
			item := map[string]interface{}{
				"id":                content.Base.Id,
				"content_source_id": content.ContentSourceId,
				"now":               now,
			}
			params["items"] = append(params["items"].([]map[string]interface{}), item)
		}
		_, err := tx.Run(ctx, `
			UNWIND $items AS item
			MERGE (n:ContentNode:Node {content_source_id: item.content_source_id})
			ON CREATE SET
				n.id = item.id,
				n.created_at = datetime(item.now),
				n.updated_at = datetime(item.now)
			ON MATCH SET
				n.updated_at = datetime(item.now)
		`, params)
		return nil, err
	})
	return err
}

// CreateClusterNodes creates or updates ClusterNodes in batch.
func (r *NodeRepo) CreateClusterNodes(ctx context.Context, clusters []*canvasv1.ClusterNode) error {
	if len(clusters) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"items": make([]map[string]interface{}, 0, len(clusters)),
		}
		now := time.Now().UTC().Format(time.RFC3339)
		for _, cluster := range clusters {
			item := map[string]interface{}{
				"id":                cluster.Base.Id,
				"space_id":          cluster.Base.SpaceId,
				"abstraction_level": cluster.Base.AbstractionLevel,
				"cluster_scope":     cluster.ClusterScope,
				"now":               now,
			}

			// Add optional fields
			if cluster.Title != nil {
				item["title"] = *cluster.Title
			}
			if len(cluster.Base.Embedding) > 0 {
				item["embedding"] = cluster.Base.Embedding
			}

			params["items"] = append(params["items"].([]map[string]interface{}), item)
		}

		_, err := tx.Run(ctx, `
			UNWIND $items AS item
			MERGE (n:ClusterNode:Node {id: item.id})
			ON CREATE SET
				n.space_id = item.space_id,
				n.abstraction_level = item.abstraction_level,
				n.cluster_scope = item.cluster_scope,
				n.created_at = datetime(item.now),
				n.updated_at = datetime(item.now)
			ON MATCH SET
				n.space_id = item.space_id,
				n.abstraction_level = item.abstraction_level,
				n.cluster_scope = item.cluster_scope,
				n.updated_at = datetime(item.now)
		`, params)

		// Handle optional fields in separate query since Neo4j UNWIND doesn't handle conditional SET well
		if len(clusters) > 0 && (clusters[0].Title != nil || len(clusters[0].Base.Embedding) > 0) {
			_, err = tx.Run(ctx, `
				UNWIND $items AS item
				MATCH (n:ClusterNode {id: item.id})
				SET n.title = CASE WHEN item.title IS NOT NULL THEN item.title ELSE n.title END,
					n.embedding = CASE WHEN item.embedding IS NOT NULL THEN item.embedding ELSE n.embedding END
			`, params)
		}

		return nil, err
	})
	return err
}

// UpdateClusterNode updates selected fields on a ClusterNode.
func (r *NodeRepo) UpdateClusterNode(ctx context.Context, update *canvasv1.ClusterNode) error {
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":  update.Base.Id,
			"now": time.Now().UTC().Format(time.RFC3339),
		}

		// Build a slice of SET clauses
		setClauses := []string{"n.updated_at = datetime($now)"}

		// Add base node clauses
		baseClauses := r.buildBaseNodeUpdateClauses(update.Base, params)
		setClauses = append(setClauses, baseClauses...)

		// Add ClusterNode-specific clauses
		if update.ClusterScope != "" {
			params["cluster_scope"] = update.ClusterScope
			setClauses = append(setClauses, "n.cluster_scope = $cluster_scope")
		}
		if update.Title != nil {
			params["title"] = *update.Title
			setClauses = append(setClauses, "n.title = $title")
		}
		if update.MemberCount != nil {
			params["member_count"] = *update.MemberCount
			setClauses = append(setClauses, "n.member_count = $member_count")
		}
		if update.CoverageScore != nil {
			params["coverage_score"] = *update.CoverageScore
			setClauses = append(setClauses, "n.coverage_score = $coverage_score")
		}

		// Join the clauses to form the final query
		query := fmt.Sprintf(
			"MATCH (n:ClusterNode {id: $id}) SET %s",
			strings.Join(setClauses, ", "),
		)

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	return err
}

// UpdateContentNode updates selected fields on a ContentNode.
func (r *NodeRepo) UpdateContentNode(ctx context.Context, update *canvasv1.ContentNode) error {
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":  update.Base.Id,
			"now": time.Now().UTC().Format(time.RFC3339),
		}

		// Build a slice of SET clauses
		setClauses := []string{"n.updated_at = datetime($now)"}

		// Add base node clauses
		baseClauses := r.buildBaseNodeUpdateClauses(update.Base, params)
		setClauses = append(setClauses, baseClauses...)

		// Add ContentNode-specific clauses
		if update.Title != nil {
			params["title"] = *update.Title
			setClauses = append(setClauses, "n.title = $title")
		}
		if update.MediaType != nil {
			params["media_type"] = *update.MediaType
			setClauses = append(setClauses, "n.media_type = $media_type")
		}
		if update.Source != nil {
			params["source"] = *update.Source
			setClauses = append(setClauses, "n.source = $source")
		}
		if update.TokenCount != nil {
			params["token_count"] = *update.TokenCount
			setClauses = append(setClauses, "n.token_count = $token_count")
		}
		// Note: ActionData handling would require protobuf struct conversion

		// Join the clauses to form the final query
		query := fmt.Sprintf(
			"MATCH (n:ContentNode {id: $id}) SET %s",
			strings.Join(setClauses, ", "),
		)

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	return err
}

// UpdateChunkNode updates selected fields on a ChunkNode.
func (r *NodeRepo) UpdateChunkNode(ctx context.Context, update *canvasv1.ChunkNode) error {
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":  update.Base.Id,
			"now": time.Now().UTC().Format(time.RFC3339),
		}

		// Build a slice of SET clauses
		setClauses := []string{"n.updated_at = datetime($now)"}

		// Add base node clauses
		baseClauses := r.buildBaseNodeUpdateClauses(update.Base, params)
		setClauses = append(setClauses, baseClauses...)

		// Add ChunkNode-specific clauses
		if update.Content != "" {
			params["content"] = update.Content
			setClauses = append(setClauses, "n.content = $content")
		}
		// Note: Cannot distinguish between "update to 0" vs "don't update" for SequenceIndex
		// since it's not an optional field (*int32) in the protobuf. If you need to set
		// SequenceIndex to 0, consider making it optional in the .proto file.
		if update.SequenceIndex != 0 {
			params["sequence_index"] = update.SequenceIndex
			setClauses = append(setClauses, "n.sequence_index = $sequence_index")
		}
		if update.Base.Position_3D != nil {
			params["location"] = map[string]interface{}{
				"x": update.Base.Position_3D.X,
				"y": update.Base.Position_3D.Y,
				"z": update.Base.Position_3D.Z,
			}
			setClauses = append(setClauses, "n.location = $location")
		}
		if update.StartPosition != nil {
			params["start_position"] = *update.StartPosition
			setClauses = append(setClauses, "n.start_position = $start_position")
		}
		if update.EndPosition != nil {
			params["end_position"] = *update.EndPosition
			setClauses = append(setClauses, "n.end_position = $end_position")
		}
		if update.ChunkType != "" {
			params["chunk_type"] = update.ChunkType
			setClauses = append(setClauses, "n.chunk_type = $chunk_type")
		}
		if update.TokenCount != nil {
			params["token_count"] = *update.TokenCount
			setClauses = append(setClauses, "n.token_count = $token_count")
		}

		// Join the clauses to form the final query
		query := fmt.Sprintf(
			"MATCH (n:ChunkNode {id: $id}) SET %s",
			strings.Join(setClauses, ", "),
		)

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	return err
}

// SoftDeleteNodes marks multiple nodes as deleted by setting deleted_at.
func (r *NodeRepo) SoftDeleteNodes(ctx context.Context, nodeIDs []string) error {
	if len(nodeIDs) == 0 {
		return nil
	}
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"ids": nodeIDs,
			"now": time.Now().UTC().Format(time.RFC3339),
		}
		_, err := tx.Run(ctx, `
			MATCH (n:Node)
			WHERE n.id IN $ids
			SET n.deleted_at = datetime($now)
		`, params)
		return nil, err
	})
	return err
}

// Build common BaseNode fields
func (r *NodeRepo) buildBaseNode(record *neo.Record) *canvasv1.BaseNode {
	base := &canvasv1.BaseNode{}

	// Required fields
	if id, ok := r.safeGetString(record, "id"); ok {
		base.Id = id
	}
	if spaceId, ok := r.safeGetString(record, "space_id"); ok {
		base.SpaceId = spaceId
	}
	if abstractionLevel, ok := r.safeGetInt64(record, "abstraction_level"); ok {
		base.AbstractionLevel = int32(abstractionLevel)
	}
	if contextType, ok := r.safeGetString(record, "context_type"); ok {
		base.ContextType = contextType
	}
	if createdAt, ok := r.safeGetTime(record, "created_at"); ok {
		base.CreatedAt = timestamppb.New(createdAt)
	}
	if updatedAt, ok := r.safeGetTime(record, "updated_at"); ok {
		base.UpdatedAt = timestamppb.New(updatedAt)
	}

	// Optional fields
	if chatContent, ok := r.safeGetString(record, "chat_content"); ok {
		base.ChatContent = &chatContent
	}
	if displayContent, ok := r.safeGetString(record, "display_content"); ok {
		base.DisplayContent = &displayContent
	}
	if semanticDensity, ok := r.safeGetFloat64(record, "semantic_density"); ok {
		base.SemanticDensity = &semanticDensity
	}
	if location, ok := r.safeGetMap(record, "location"); ok {
		if x, ok := location["x"].(float64); ok {
			y := location["y"].(float64)
			z := location["z"].(float64)
			base.Position_3D = &canvasv1.SpatialCoordinates{
				X: int32(x),
				Y: int32(y),
				Z: int32(z),
			}
		}
	}

	// Array fields
	if keywords, ok := r.safeGetStringSlice(record, "keywords"); ok {
		base.Keywords = keywords
	}
	if embedding, ok := r.safeGetFloatSlice(record, "embedding"); ok {
		base.Embedding = embedding
	}

	return base
}

// Helper function to build SET clauses for BaseNode fields
func (r *NodeRepo) buildBaseNodeUpdateClauses(base *canvasv1.BaseNode, params map[string]interface{}) []string {
	var clauses []string

	if len(base.Embedding) > 0 {
		params["embedding"] = base.Embedding
		clauses = append(clauses, "n.embedding = $embedding")
	}

	if base.MlInfo != nil {
		if base.MlInfo.EmbeddingModelId != "" {
			params["model_id"] = base.MlInfo.EmbeddingModelId
			clauses = append(clauses, "n.model_id = $model_id")
		}
		if base.MlInfo.ModelVersion != "" {
			params["model_version"] = base.MlInfo.ModelVersion
			clauses = append(clauses, "n.model_version = $model_version")
		}
	}

	return clauses
}

// Helper methods to build specific node types from database records
func (r *NodeRepo) buildContentNode(record *neo.Record) *canvasv1.Node {
	// Build common base fields
	base := r.buildBaseNode(record)

	// Get ContentNode specific fields
	contentSourceId, _ := r.safeGetString(record, "content_source_id")
	if contentSourceId == "" {
		return nil // Required field missing
	}

	// Build ContentNode
	contentNode := &canvasv1.ContentNode{
		Base:            base,
		ContentSourceId: contentSourceId,
	}

	// Add optional fields if they exist
	if title, ok := r.safeGetString(record, "title"); ok {
		contentNode.Title = &title
	}
	if mediaType, ok := r.safeGetString(record, "media_type"); ok {
		contentNode.MediaType = &mediaType
	}
	if source, ok := r.safeGetString(record, "source"); ok {
		contentNode.Source = &source
	}
	if tokenCount, ok := r.safeGetInt64(record, "token_count"); ok {
		val := int32(tokenCount)
		contentNode.TokenCount = &val
	}

	// Wrap in Node
	return &canvasv1.Node{
		Node: &canvasv1.Node_Content{
			Content: contentNode,
		},
	}
}

func (r *NodeRepo) buildChunkNode(record *neo.Record) *canvasv1.Node {
	// Build common base fields
	base := r.buildBaseNode(record)

	// Get ChunkNode specific fields
	contentSourceId, _ := r.safeGetString(record, "content_source_id")
	if contentSourceId == "" {
		return nil // Required field missing
	}

	sequenceIndex, _ := r.safeGetInt64(record, "sequence_index")
	content, _ := r.safeGetString(record, "content")
	if content == "" {
		return nil // Required field missing
	}

	// Build ChunkNode
	chunkNode := &canvasv1.ChunkNode{
		Base:            base,
		ContentSourceId: contentSourceId,
		SequenceIndex:   int32(sequenceIndex),
		Content:         content,
	}

	// Add optional fields if they exist
	if chunkType, ok := r.safeGetString(record, "chunk_type"); ok {
		chunkNode.ChunkType = chunkType
	}
	if startPosition, ok := r.safeGetInt64(record, "start_position"); ok {
		chunkNode.StartPosition = &startPosition
	}
	if endPosition, ok := r.safeGetInt64(record, "end_position"); ok {
		chunkNode.EndPosition = &endPosition
	}
	if tokenCount, ok := r.safeGetInt64(record, "token_count"); ok {
		val := int32(tokenCount)
		chunkNode.TokenCount = &val
	}

	// Wrap in Node
	return &canvasv1.Node{
		Node: &canvasv1.Node_Chunk{
			Chunk: chunkNode,
		},
	}
}

func (r *NodeRepo) buildClusterNode(record *neo.Record) *canvasv1.Node {
	// Build common base fields
	base := r.buildBaseNode(record)

	// Get ClusterNode specific fields
	clusterScope, _ := r.safeGetString(record, "cluster_scope")
	if clusterScope == "" {
		return nil // Required field missing
	}

	// Build ClusterNode
	clusterNode := &canvasv1.ClusterNode{
		Base:         base,
		ClusterScope: clusterScope,
	}

	// Add optional fields if they exist
	if title, ok := r.safeGetString(record, "title"); ok {
		clusterNode.Title = &title
	}
	if memberCount, ok := r.safeGetInt64(record, "member_count"); ok {
		val := int32(memberCount)
		clusterNode.MemberCount = &val
	}
	if coverageScore, ok := r.safeGetFloat64(record, "coverage_score"); ok {
		clusterNode.CoverageScore = &coverageScore
	}

	// Wrap in Node
	return &canvasv1.Node{
		Node: &canvasv1.Node_Cluster{
			Cluster: clusterNode,
		},
	}
}

// Safe getter helpers to prevent panics from unsafe type assertions
func (r *NodeRepo) safeGetString(record *neo.Record, key string) (string, bool) {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return "", false
	}
	strVal, ok := val.(string)
	return strVal, ok
}

func (r *NodeRepo) safeGetInt64(record *neo.Record, key string) (int64, bool) {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return 0, false
	}
	intVal, ok := val.(int64)
	return intVal, ok
}

func (r *NodeRepo) safeGetFloat64(record *neo.Record, key string) (float64, bool) {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return 0, false
	}
	floatVal, ok := val.(float64)
	return floatVal, ok
}

func (r *NodeRepo) safeGetTime(record *neo.Record, key string) (time.Time, bool) {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return time.Time{}, false
	}
	timeVal, ok := val.(time.Time)
	return timeVal, ok
}

func (r *NodeRepo) safeGetMap(record *neo.Record, key string) (map[string]interface{}, bool) {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return nil, false
	}
	mapVal, ok := val.(map[string]interface{})
	return mapVal, ok
}

func (r *NodeRepo) safeGetStringSlice(record *neo.Record, key string) ([]string, bool) {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return nil, false
	}
	slice, ok := val.([]interface{})
	if !ok {
		return nil, false
	}
	result := make([]string, len(slice))
	for i, v := range slice {
		if strVal, ok := v.(string); ok {
			result[i] = strVal
		} else {
			return nil, false
		}
	}
	return result, true
}

func (r *NodeRepo) safeGetFloatSlice(record *neo.Record, key string) ([]float32, bool) {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return nil, false
	}
	slice, ok := val.([]interface{})
	if !ok {
		return nil, false
	}
	result := make([]float32, len(slice))
	for i, v := range slice {
		if floatVal, ok := v.(float64); ok {
			result[i] = float32(floatVal)
		} else {
			return nil, false
		}
	}
	return result, true
}
