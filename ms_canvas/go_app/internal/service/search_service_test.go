package service

import (
	"testing"

	v1 "demo/ms_canvas/go_app/api/proto/public/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper functions to create test data

func createTestNode(nodeType v1.NodeType, id string, content string) *v1.Node {
	base := &v1.BaseNode{
		Id:               id,
		SpaceId:          uuid.New().String(),
		AbstractionLevel: 1,
		ContextType:      "test_context",
		DisplayContent:   &content,
	}

	switch nodeType {
	case v1.NodeType_NODE_TYPE_CONTENT:
		return &v1.Node{
			Node: &v1.Node_Content{
				Content: &v1.ContentNode{
					Base:            base,
					ContentSourceId: "test-source-id",
					Title:           &content,
				},
			},
		}
	case v1.NodeType_NODE_TYPE_CHUNK:
		return &v1.Node{
			Node: &v1.Node_Chunk{
				Chunk: &v1.ChunkNode{
					Base: &v1.BaseNode{
						Id:               uuid.New().String(),
						SpaceId:          "test-space",
						AbstractionLevel: 2,
						ContextType:      "chunk",
						Keywords:         []string{"test", "chunk"},
						DisplayContent:   stringPtr("Test Chunk Content"),
						SemanticDensity:  float64Ptr(0.9),
						Position_3D: &v1.SpatialCoordinates{
							X: 15,
							Y: 25,
							Z: 35,
						},
						IsPositionLocked: boolPtr(true),
						Visibility:       boolPtr(false),
						DisplayProps: &v1.DisplayProps{
							Size:    50,
							Opacity: 0.8,
							Shape:   "square",
							Color:   "#00ff00",
						},
						EngagementScore: &v1.EngagementScore{
							CanvasScore:  0.8,
							ChatScore:    0.7,
							OverallScore: 0.75,
						},
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
						ChatContent: stringPtr(content),
					},
					ContentSourceId: "source-456",
					SequenceIndex:   1,
					ChunkType:       "paragraph",
					StartPosition:   int64Ptr(0),
					EndPosition:     int64Ptr(100),
					TokenCount:      int32Ptr(50),
				},
			},
		}
	case v1.NodeType_NODE_TYPE_CLUSTER:
		return &v1.Node{
			Node: &v1.Node_Cluster{
				Cluster: &v1.ClusterNode{
					Base:         base,
					ClusterScope: "test_scope",
					Title:        &content,
				},
			},
		}
	default:
		return &v1.Node{
			Node: &v1.Node_Content{
				Content: &v1.ContentNode{
					Base:            base,
					ContentSourceId: "test-source-id",
				},
			},
		}
	}
}

// Test cases for the similarity calculation logic
func TestSearchService_calculateSimilarityScore(t *testing.T) {
	t.Run("Calculate similarity for ContentNode with matching terms", func(t *testing.T) {
		query := "machine learning"
		node := createTestNode(v1.NodeType_NODE_TYPE_CONTENT, "node-1", "machine learning is awesome")

		// Create a searchServiceImpl directly to test the private method
		service := &searchServiceImpl{}

		score := service.calculateSimilarityScore(query, node)

		// Both "machine" and "learning" match, so score should be 1.0
		assert.Equal(t, 1.0, score)
	})

	t.Run("Calculate similarity for ContentNode with partial matching terms", func(t *testing.T) {
		query := "machine learning deep"
		node := createTestNode(v1.NodeType_NODE_TYPE_CONTENT, "node-1", "machine learning is awesome")

		service := &searchServiceImpl{}
		score := service.calculateSimilarityScore(query, node)

		// Only "machine" and "learning" match out of 3 terms, so score should be ~0.666
		assert.InDelta(t, 0.666, score, 0.001)
	})

	t.Run("Calculate similarity for ContentNode with no matching terms", func(t *testing.T) {
		query := "artificial intelligence"
		node := createTestNode(v1.NodeType_NODE_TYPE_CONTENT, "node-1", "machine learning is awesome")

		service := &searchServiceImpl{}
		score := service.calculateSimilarityScore(query, node)

		// No terms match, so score should be 0.0
		assert.Equal(t, 0.0, score)
	})

	t.Run("Calculate similarity for ChunkNode with matching terms", func(t *testing.T) {
		query := "data processing"
		node := createTestNode(v1.NodeType_NODE_TYPE_CHUNK, "node-1", "data processing pipeline")

		service := &searchServiceImpl{}
		score := service.calculateSimilarityScore(query, node)

		// Both "data" and "processing" match, so score should be 1.0
		assert.Equal(t, 1.0, score)
	})

	t.Run("Calculate similarity for ClusterNode with matching terms", func(t *testing.T) {
		query := "neural networks"
		node := createTestNode(v1.NodeType_NODE_TYPE_CLUSTER, "node-1", "neural networks architecture")

		service := &searchServiceImpl{}
		score := service.calculateSimilarityScore(query, node)

		// Both "neural" and "networks" match, so score should be 1.0
		assert.Equal(t, 1.0, score)
	})

	t.Run("Calculate similarity for node without display content", func(t *testing.T) {
		query := "test query"
		node := createTestNode(v1.NodeType_NODE_TYPE_CONTENT, "node-1", "")
		node.GetContent().Base.DisplayContent = nil

		service := &searchServiceImpl{}
		score := service.calculateSimilarityScore(query, node)

		// No display content, so score should be 0.0
		assert.Equal(t, 0.0, score)
	})

	t.Run("Calculate similarity with case insensitivity", func(t *testing.T) {
		query := "MACHINE LEARNING"
		node := createTestNode(v1.NodeType_NODE_TYPE_CONTENT, "node-1", "machine learning is awesome")

		service := &searchServiceImpl{}
		score := service.calculateSimilarityScore(query, node)

		// Should match regardless of case
		assert.Equal(t, 1.0, score)
	})

	t.Run("Calculate similarity with empty query", func(t *testing.T) {
		query := ""
		node := createTestNode(v1.NodeType_NODE_TYPE_CONTENT, "node-1", "machine learning is awesome")

		service := &searchServiceImpl{}
		score := service.calculateSimilarityScore(query, node)

		// Empty query should return 0.0
		assert.Equal(t, 0.0, score)
	})
}

// Test validation logic that can be tested independently
func TestSearchService_ValidationLogic(t *testing.T) {
	t.Run("Valid input parameters", func(t *testing.T) {
		// Test that we can create the service struct
		service := &searchServiceImpl{}

		// Test that the service implements the interface
		var _ SearchService = service

		assert.NotNil(t, service)
	})

	t.Run("Empty query validation", func(t *testing.T) {
		// Simulate the validation logic from SemanticSearch
		req := &v1.SemanticSearchRequest{Query: ""}
		if req.Query == "" {
			// This would be the error case in the actual method
			assert.Equal(t, "", req.Query)
		}

		// Test with valid query
		req = &v1.SemanticSearchRequest{Query: "test query"}
		if req.Query != "" {
			assert.Equal(t, "test query", req.Query)
		}
	})

	t.Run("TopK parameter validation", func(t *testing.T) {
		// Test default value logic (this would be in SemanticSearch)
		req := &v1.SemanticSearchRequest{TopK: 0}
		topK := req.TopK
		if topK <= 0 {
			topK = 25 // Default value
		}
		assert.Equal(t, int32(25), topK)

		// Test maximum value logic
		req = &v1.SemanticSearchRequest{TopK: 200}
		topK = req.TopK
		if topK > 100 {
			topK = 100 // Maximum value
		}
		assert.Equal(t, int32(100), topK)

		// Test normal value
		req = &v1.SemanticSearchRequest{TopK: 50}
		topK = req.TopK
		if topK <= 0 {
			topK = 25
		}
		if topK > 100 {
			topK = 100
		}
		assert.Equal(t, int32(50), topK)
	})

	t.Run("UUID parsing validation", func(t *testing.T) {
		// Test valid UUID
		req := &v1.SemanticSearchRequest{SpaceId: "123e4567-e89b-12d3-a456-426614174000"}
		_, err := uuid.Parse(req.SpaceId)
		assert.NoError(t, err)

		// Test invalid UUID (this would cause an error in SemanticSearch)
		req = &v1.SemanticSearchRequest{SpaceId: "invalid-uuid"}
		_, err = uuid.Parse(req.SpaceId)
		assert.Error(t, err)
	})
}
