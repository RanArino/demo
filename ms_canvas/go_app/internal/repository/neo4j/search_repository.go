package neo4j

import (
	"context"
	"fmt"
	"strings"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"

	"github.com/google/uuid"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type SearchRepo struct {
	driver *Driver
}

func NewSearchRepo(driver *Driver) *SearchRepo {
	return &SearchRepo{driver: driver}
}

// VectorSearch performs semantic search using Neo4j's native vector search
// If nodeTypes is empty or nil, searches across all node types (ClusterNode, ContentNode, ChunkNode)
// using their respective vector indexes. If nodeTypes is specified, only searches the specified types.
func (r *SearchRepo) VectorSearch(ctx context.Context, spaceID uuid.UUID, queryEmbedding []float32, topK int32, nodeTypes []canvasv1.NodeType) ([]*canvasv1.Node, error) {
	sess := r.driver.NewSession(ctx, neo.SessionConfig{DatabaseName: r.driver.dbName})
	defer sess.Close(ctx)

	// Convert node types to labels for filtering (only if nodeTypes specified)
	var labels []string
	if len(nodeTypes) > 0 {
		labels = make([]string, len(nodeTypes))
		for i, nt := range nodeTypes {
			switch nt {
			case canvasv1.NodeType_NODE_TYPE_CLUSTER:
				labels[i] = "ClusterNode"
			case canvasv1.NodeType_NODE_TYPE_CONTENT:
				labels[i] = "ContentNode"
			case canvasv1.NodeType_NODE_TYPE_CHUNK:
				labels[i] = "ChunkNode"
			}
		}
	}

	result, err := sess.ExecuteRead(ctx, func(tx neo.ManagedTransaction) (interface{}, error) {
		// Build dynamic query based on node types
		var queryParts []string

		// If no nodeTypes specified, search across all nodes using common Node label
		if len(nodeTypes) == 0 {
			// Search all three indexes and combine results
			queryParts = append(queryParts, `
				// Search ClusterNode index
				CALL db.index.vector.queryNodes('clusternode_embedding', $topK, $queryEmbedding)
				YIELD node AS clusterNode, score AS clusterScore
				WHERE clusterNode.space_id = $spaceID
				RETURN clusterNode, clusterScore, 'ClusterNode' as nodeType

				UNION ALL

				// Search ContentNode index
				CALL db.index.vector.queryNodes('contentnode_embedding', $topK, $queryEmbedding)
				YIELD node AS contentNode, score AS contentScore
				WHERE contentNode.space_id = $spaceID
				RETURN contentNode, contentScore, 'ContentNode' as nodeType

				UNION ALL

				// Search ChunkNode index
				CALL db.index.vector.queryNodes('chunknode_embedding', $topK, $queryEmbedding)
				YIELD node AS chunkNode, score AS chunkScore
				WHERE chunkNode.space_id = $spaceID
				RETURN chunkNode, chunkScore, 'ChunkNode' as nodeType
			`)
		} else {
			// Specific node types requested - search only requested types
			for _, nodeType := range nodeTypes {
				switch nodeType {
				case canvasv1.NodeType_NODE_TYPE_CLUSTER:
					queryParts = append(queryParts, `
						CALL db.index.vector.queryNodes('clusternode_embedding', $topK, $queryEmbedding)
						YIELD node AS clusterNode, score AS clusterScore
						WHERE clusterNode.space_id = $spaceID AND 'ClusterNode' IN $labels
						RETURN clusterNode, clusterScore, 'ClusterNode' as nodeType
					`)
				case canvasv1.NodeType_NODE_TYPE_CONTENT:
					queryParts = append(queryParts, `
						CALL db.index.vector.queryNodes('contentnode_embedding', $topK, $queryEmbedding)
						YIELD node AS contentNode, score AS contentScore
						WHERE contentNode.space_id = $spaceID AND 'ContentNode' IN $labels
						RETURN contentNode, contentScore, 'ContentNode' as nodeType
					`)
				case canvasv1.NodeType_NODE_TYPE_CHUNK:
					queryParts = append(queryParts, `
						CALL db.index.vector.queryNodes('chunknode_embedding', $topK, $queryEmbedding)
						YIELD node AS chunkNode, score AS chunkScore
						WHERE chunkNode.space_id = $spaceID AND 'ChunkNode' IN $labels
						RETURN chunkNode, chunkScore, 'ChunkNode' as nodeType
					`)
				}
			}

			if len(queryParts) == 0 {
				return []*canvasv1.Node{}, nil
			}
		}

		query := strings.Join(queryParts, " UNION ALL ") + `
			ORDER BY score DESC
			LIMIT $topK
		`

		// Prepare parameters - only include labels if nodeTypes were specified
		params := map[string]interface{}{
			"queryEmbedding": queryEmbedding,
			"spaceID":        spaceID.String(),
			"topK":           topK,
		}

		if len(nodeTypes) > 0 {
			params["labels"] = labels
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, fmt.Errorf("failed to execute vector search query: %w", err)
		}

		var nodes []*canvasv1.Node
		for result.Next(ctx) {
			record := result.Record()

			// Get the node and score
			nodeInterface, _ := record.Get("node")
			score, _ := record.Get("score")
			nodeType, _ := record.Get("nodeType")

			// Convert Neo4j node to our Node struct with score
			node := r.convertNeo4jNodeToProtobufWithScore(nodeInterface, score.(float64), nodeType.(string))
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

// MultiHopSearch performs real-time multi-hop search across abstraction levels
// Returns results at each level immediately for progressive user interaction
// DRAFT: Basic implementation - will be enhanced later
func (r *SearchRepo) MultiHopSearch(ctx context.Context, spaceID uuid.UUID, queryEmbedding []float32, topK int32) (*MultiHopSearchResponse, error) {
	// TODO: Implement proper response structure for multi-hop results
	// This should return results at each abstraction level immediately

	// For now, return basic structure - will be implemented properly later
	response := &MultiHopSearchResponse{
		ClusterResults: []*canvasv1.Node{},
		ContentResults: []*canvasv1.Node{},
		ChunkResults:   []*canvasv1.Node{},
	}

	// DRAFT: Search each level separately and return immediately
	// This provides real-time results at each abstraction level

	// Step 1: Search ClusterNodes
	if clusters, err := r.VectorSearch(ctx, spaceID, queryEmbedding, topK, []canvasv1.NodeType{canvasv1.NodeType_NODE_TYPE_CLUSTER}); err == nil {
		response.ClusterResults = clusters
	}

	// Step 2: Search ContentNodes
	if contents, err := r.VectorSearch(ctx, spaceID, queryEmbedding, topK, []canvasv1.NodeType{canvasv1.NodeType_NODE_TYPE_CONTENT}); err == nil {
		response.ContentResults = contents
	}

	// Step 3: Search ChunkNodes
	if chunks, err := r.VectorSearch(ctx, spaceID, queryEmbedding, topK, []canvasv1.NodeType{canvasv1.NodeType_NODE_TYPE_CHUNK}); err == nil {
		response.ChunkResults = chunks
	}

	return response, nil
}

// MultiHopSearchResponse contains results from each abstraction level
type MultiHopSearchResponse struct {
	ClusterResults []*canvasv1.Node `json:"cluster_results"`
	ContentResults []*canvasv1.Node `json:"content_results"`
	ChunkResults   []*canvasv1.Node `json:"chunk_results"`
}

// convertNeo4jNodeToProtobufWithScore converts a Neo4j node to protobuf format with similarity score
func (r *SearchRepo) convertNeo4jNodeToProtobufWithScore(nodeInterface interface{}, score float64, nodeType string) *canvasv1.Node {
	// Convert the node interface to Neo4j node
	neoNode := nodeInterface.(neo.Node)

	// Extract properties
	props := neoNode.Props
	id, _ := props["id"].(string)
	spaceId, _ := props["space_id"].(string)
	abstractionLevel, _ := props["abstraction_level"].(int64)
	contextType, _ := props["context_type"].(string)

	// Build base node
	base := &canvasv1.BaseNode{
		Id:               id,
		SpaceId:          spaceId,
		AbstractionLevel: int32(abstractionLevel),
		ContextType:      contextType,
	}

	// Add optional fields if they exist
	if chatContent, ok := props["chat_content"].(string); ok {
		base.ChatContent = &chatContent
	}
	if displayContent, ok := props["display_content"].(string); ok {
		base.DisplayContent = &displayContent
	}
	if semanticDensity, ok := props["semantic_density"].(float64); ok {
		base.SemanticDensity = &semanticDensity
	}
	if keywords, ok := props["keywords"].([]interface{}); ok {
		base.Keywords = make([]string, len(keywords))
		for i, k := range keywords {
			if str, ok := k.(string); ok {
				base.Keywords[i] = str
			}
		}
	}
	if embedding, ok := props["embedding"].([]interface{}); ok {
		base.Embedding = make([]float32, len(embedding))
		for i, e := range embedding {
			if f, ok := e.(float64); ok {
				base.Embedding[i] = float32(f)
			}
		}
	}

	// Build specific node type
	var node *canvasv1.Node
	switch nodeType {
	case "ContentNode":
		contentNode := &canvasv1.ContentNode{
			Base:            base,
			ContentSourceId: props["content_source_id"].(string),
		}
		if title, ok := props["title"].(string); ok {
			contentNode.Title = &title
		}
		if mediaType, ok := props["media_type"].(string); ok {
			contentNode.MediaType = &mediaType
		}
		if source, ok := props["source"].(string); ok {
			contentNode.Source = &source
		}
		if tokenCount, ok := props["token_count"].(int64); ok {
			val := int32(tokenCount)
			contentNode.TokenCount = &val
		}
		node = &canvasv1.Node{
			Node: &canvasv1.Node_Content{
				Content: contentNode,
			},
		}

	case "ChunkNode":
		chunkNode := &canvasv1.ChunkNode{
			Base:            base,
			ContentSourceId: props["content_source_id"].(string),
			SequenceIndex:   int32(props["sequence_index"].(int64)),
			Content:         props["content"].(string),
		}
		if chunkType, ok := props["chunk_type"].(string); ok {
			chunkNode.ChunkType = chunkType
		}
		if tokenCount, ok := props["token_count"].(int64); ok {
			val := int32(tokenCount)
			chunkNode.TokenCount = &val
		}
		node = &canvasv1.Node{
			Node: &canvasv1.Node_Chunk{
				Chunk: chunkNode,
			},
		}

	case "ClusterNode":
		clusterNode := &canvasv1.ClusterNode{
			Base:         base,
			ClusterScope: props["cluster_scope"].(string),
		}
		if title, ok := props["title"].(string); ok {
			clusterNode.Title = &title
		}
		if memberCount, ok := props["member_count"].(int64); ok {
			val := int32(memberCount)
			clusterNode.MemberCount = &val
		}
		if coverageScore, ok := props["coverage_score"].(float64); ok {
			clusterNode.CoverageScore = &coverageScore
		}
		node = &canvasv1.Node{
			Node: &canvasv1.Node_Cluster{
				Cluster: clusterNode,
			},
		}
	}

	return node
}
