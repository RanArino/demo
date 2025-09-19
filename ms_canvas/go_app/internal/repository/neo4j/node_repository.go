package neo4j

import (
	"context"
	"time"

	"demo/ms_canvas/go_app/internal/domain"

	"github.com/google/uuid"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NodeRepo struct {
	driver *Driver
}

func NewNodeRepo(driver *Driver) *NodeRepo { return &NodeRepo{driver: driver} }

// CreateChunkNodes creates chunk nodes (idempotent on id/content_source_id via MERGE).
func (r *NodeRepo) CreateChunkNodes(chunks []domain.ChunkNode) error {
	if len(chunks) == 0 {
		return nil
	}
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"items": make([]map[string]interface{}, 0, len(chunks)),
		}
		now := time.Now().UTC().Format(time.RFC3339)
		for _, c := range chunks {
			id := c.ID.String()
			params["items"] = append(params["items"].([]map[string]interface{}), map[string]interface{}{
				"id":                id,
				"content_source_id": c.ContentSourceID.String(),
				"sequence_index":    c.SequenceIndex,
				"location": map[string]interface{}{
					"x": c.Position3D.X,
					"y": c.Position3D.Y,
					"z": c.Position3D.Z,
				},
				"start_position": c.StartPosition,
				"end_position":   c.EndPosition,
				"content":        c.Content,
				"now":            now,
			})
		}
		_, err := tx.Run(ctx, `
			UNWIND $items AS item
			MERGE (n:ChunkNode {id: item.id})
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

// CreateContentNode creates or updates a ContentNode for the given content_source_id.
func (r *NodeRepo) CreateContentNode(node domain.ContentNode) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"content_source_id": node.ContentSourceID.String(),
			"now":               time.Now().UTC().Format(time.RFC3339),
			"id":                node.ID.String(),
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

// CreateClusterNode creates or updates a ClusterNode with minimal required fields.
func (r *NodeRepo) CreateClusterNode(id uuid.UUID, spaceID uuid.UUID, abstractionLevel int, clusterScope string, title *string, embedding *[]float32) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":                id.String(),
			"space_id":          spaceID.String(),
			"abstraction_level": abstractionLevel,
			"cluster_scope":     clusterScope,
			"now":               time.Now().UTC().Format(time.RFC3339),
		}

		setCreate := "n.space_id = $space_id, n.abstraction_level = $abstraction_level, n.cluster_scope = $cluster_scope"
		setMatch := "n.space_id = $space_id, n.abstraction_level = $abstraction_level, n.cluster_scope = $cluster_scope"

		if title != nil {
			params["title"] = *title
			setCreate += ", n.title = $title"
			setMatch += ", n.title = $title"
		}
		if embedding != nil {
			params["embedding"] = *embedding
			setCreate += ", n.embedding = $embedding"
			setMatch += ", n.embedding = $embedding"
		}

		// Always set timestamps
		setCreate += ", n.created_at = datetime($now), n.updated_at = datetime($now)"
		setMatch += ", n.updated_at = datetime($now)"

		query := "MERGE (n:ClusterNode {id: $id}) ON CREATE SET " + setCreate + " ON MATCH SET " + setMatch

		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	return err
}

// UpdateClusterNode updates selected fields on a ClusterNode.
func (r *NodeRepo) UpdateClusterNode(update domain.ClusterNodeUpdate) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":  update.ID.String(),
			"now": time.Now().UTC().Format(time.RFC3339),
		}
		set := "n.updated_at = datetime($now)"
		if update.BaseNodeUpdate.Embedding != nil {
			params["embedding"] = *update.BaseNodeUpdate.Embedding
			set += ", n.embedding = $embedding"
		}
		if update.BaseNodeUpdate.ML != nil && update.BaseNodeUpdate.ML.EmbeddingModelID != nil {
			params["model_id"] = *update.BaseNodeUpdate.ML.EmbeddingModelID
			set += ", n.model_id = $model_id"
		}
		if update.BaseNodeUpdate.ML != nil && update.BaseNodeUpdate.ML.ModelVersion != nil {
			params["model_version"] = *update.BaseNodeUpdate.ML.ModelVersion
			set += ", n.model_version = $model_version"
		}
		if update.ClusterScope != nil {
			params["cluster_scope"] = *update.ClusterScope
			set += ", n.cluster_scope = $cluster_scope"
		}
		if update.Title != nil {
			params["title"] = *update.Title
			set += ", n.title = $title"
		}
		if update.MemberCount != nil {
			params["member_count"] = *update.MemberCount
			set += ", n.member_count = $member_count"
		}
		if update.CoverageScore != nil {
			params["coverage_score"] = *update.CoverageScore
			set += ", n.coverage_score = $coverage_score"
		}
		query := "MATCH (n:ClusterNode {id: $id}) SET " + set
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	return err
}

// UpdateContentNode updates selected fields on a ContentNode.
func (r *NodeRepo) UpdateContentNode(update domain.ContentNodeUpdate) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":  update.ID.String(),
			"now": time.Now().UTC().Format(time.RFC3339),
		}
		set := "n.updated_at = datetime($now)"
		if update.BaseNodeUpdate.Embedding != nil {
			params["embedding"] = *update.BaseNodeUpdate.Embedding
			set += ", n.embedding = $embedding"
		}
		if update.BaseNodeUpdate.ML != nil && update.BaseNodeUpdate.ML.EmbeddingModelID != nil {
			params["model_id"] = *update.BaseNodeUpdate.ML.EmbeddingModelID
			set += ", n.model_id = $model_id"
		}
		if update.BaseNodeUpdate.ML != nil && update.BaseNodeUpdate.ML.ModelVersion != nil {
			params["model_version"] = *update.BaseNodeUpdate.ML.ModelVersion
			set += ", n.model_version = $model_version"
		}
		if update.Title != nil {
			params["title"] = *update.Title
			set += ", n.title = $title"
		}
		if update.MediaType != nil {
			params["media_type"] = *update.MediaType
			set += ", n.media_type = $media_type"
		}
		if update.Source != nil {
			params["source"] = *update.Source
			set += ", n.source = $source"
		}
		if update.TokenCount != nil {
			params["token_count"] = *update.TokenCount
			set += ", n.token_count = $token_count"
		}
		if update.ActionData != nil {
			params["action_data"] = *update.ActionData
			set += ", n.action_data = $action_data"
		}
		query := "MATCH (n:ContentNode {id: $id}) SET " + set
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	return err
}

// UpdateChunkNode updates selected fields on a ChunkNode.
func (r *NodeRepo) UpdateChunkNode(update domain.ChunkNodeUpdate) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":  update.ID.String(),
			"now": time.Now().UTC().Format(time.RFC3339),
		}
		set := "n.updated_at = datetime($now)"
		if update.BaseNodeUpdate.Embedding != nil {
			params["embedding"] = *update.BaseNodeUpdate.Embedding
			set += ", n.embedding = $embedding"
		}
		if update.BaseNodeUpdate.ML != nil && update.BaseNodeUpdate.ML.EmbeddingModelID != nil {
			params["model_id"] = *update.BaseNodeUpdate.ML.EmbeddingModelID
			set += ", n.model_id = $model_id"
		}
		if update.BaseNodeUpdate.ML != nil && update.BaseNodeUpdate.ML.ModelVersion != nil {
			params["model_version"] = *update.BaseNodeUpdate.ML.ModelVersion
			set += ", n.model_version = $model_version"
		}
		if update.Content != nil {
			params["content"] = *update.Content
			set += ", n.content = $content"
		}
		if update.SequenceIndex != nil {
			params["sequence_index"] = *update.SequenceIndex
			set += ", n.sequence_index = $sequence_index"
		}
		if update.Position3D != nil {
			params["location"] = map[string]interface{}{"x": update.Position3D.X, "y": update.Position3D.Y, "z": update.Position3D.Z}
			set += ", n.location = $location"
		}
		if update.StartPosition != nil {
			params["start_position"] = *update.StartPosition
			set += ", n.start_position = $start_position"
		}
		if update.EndPosition != nil {
			params["end_position"] = *update.EndPosition
			set += ", n.end_position = $end_position"
		}
		if update.ChunkType != nil {
			params["chunk_type"] = *update.ChunkType
			set += ", n.chunk_type = $chunk_type"
		}
		if update.TokenCount != nil {
			params["token_count"] = *update.TokenCount
			set += ", n.token_count = $token_count"
		}
		query := "MATCH (n:ChunkNode {id: $id}) SET " + set
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	return err
}

// SoftDeleteNode marks a node as deleted by setting deleted_at.
func (r *NodeRepo) SoftDeleteNode(nodeID uuid.UUID) error {
	ctx := context.Background()
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":  nodeID.String(),
			"now": time.Now().UTC().Format(time.RFC3339),
		}
		_, err := tx.Run(ctx, `
			MATCH (n {id: $id})
			SET n.deleted_at = datetime($now)
		`, params)
		return nil, err
	})
	return err
}
