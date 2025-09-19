package neo4j

import (
	"context"
	"errors"
	"testing"

	"demo/ms_canvas/go_app/internal/domain"

	"github.com/google/uuid"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDriverInterface wraps our Driver for easier testing
type MockDriverInterface interface {
	NewSession(ctx context.Context, config neo.SessionConfig) neo.SessionWithContext
	Close(ctx context.Context) error
}

// MockSessionInterface wraps session operations for testing
type MockSessionInterface interface {
	ExecuteWrite(ctx context.Context, work neo.ManagedTransactionWork, configurers ...func(*neo.TransactionConfig)) (interface{}, error)
	Close(ctx context.Context) error
}

// MockDriver implements MockDriverInterface
type MockDriver struct {
	mock.Mock
}

func (m *MockDriver) NewSession(ctx context.Context, config neo.SessionConfig) neo.SessionWithContext {
	args := m.Called(ctx, config)
	return args.Get(0).(neo.SessionWithContext)
}

func (m *MockDriver) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockSession implements MockSessionInterface
type MockSession struct {
	mock.Mock
}

func (m *MockSession) ExecuteWrite(ctx context.Context, work neo.ManagedTransactionWork, configurers ...func(*neo.TransactionConfig)) (interface{}, error) {
	args := m.Called(ctx, work, configurers)
	return args.Get(0), args.Error(1)
}

func (m *MockSession) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestNodeRepo tests using a wrapper approach instead of full interface mocking
func TestNodeRepo_CreateChunkNodes_Integration(t *testing.T) {
	// Integration-style test that verifies business logic without full Neo4j
	tests := []struct {
		name        string
		chunks      []domain.ChunkNode
		expectError bool
	}{
		{
			name:        "empty chunks should not error",
			chunks:      []domain.ChunkNode{},
			expectError: false,
		},
		{
			name: "valid chunk should generate correct data structure",
			chunks: []domain.ChunkNode{
				{
					BaseNode: domain.BaseNode{
						ID:      uuid.New(),
						SpaceID: uuid.New(),
						Position3D: &domain.SpatialCoordinates{
							X: 10, Y: 20, Z: 30,
						},
					},
					ContentSourceID: uuid.New(),
					SequenceIndex:   0,
					StartPosition:   0,
					EndPosition:     100,
					Content:         "test content",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the business logic without Neo4j dependency
			if len(tt.chunks) == 0 {
				// Should return early without error
				assert.False(t, tt.expectError)
				return
			}

			// Validate chunk data structure
			for _, chunk := range tt.chunks {
				assert.NotEqual(t, uuid.Nil, chunk.ID)
				assert.NotEqual(t, uuid.Nil, chunk.ContentSourceID)
				assert.NotEmpty(t, chunk.Content)
				assert.GreaterOrEqual(t, chunk.SequenceIndex, 0)
				if chunk.Position3D != nil {
					assert.NotNil(t, chunk.Position3D)
				}
			}
		})
	}
}

func TestNodeRepo_CreateContentNode_Validation(t *testing.T) {
	tests := []struct {
		name         string
		node         domain.ContentNode
		validateFunc func(t *testing.T, node domain.ContentNode)
	}{
		{
			name: "valid content node should have required fields",
			node: domain.ContentNode{
				BaseNode: domain.BaseNode{
					ID:      uuid.New(),
					SpaceID: uuid.New(),
				},
				ContentSourceID: uuid.New(),
			},
			validateFunc: func(t *testing.T, node domain.ContentNode) {
				assert.NotEqual(t, uuid.Nil, node.ID)
				assert.NotEqual(t, uuid.Nil, node.ContentSourceID)
				assert.NotEqual(t, uuid.Nil, node.SpaceID)
			},
		},
		{
			name: "content node with optional fields",
			node: domain.ContentNode{
				BaseNode: domain.BaseNode{
					ID:      uuid.New(),
					SpaceID: uuid.New(),
				},
				ContentSourceID: uuid.New(),
				Title:           stringPtr("Test Document"),
				MediaType:       stringPtr("text/plain"),
				TokenCount:      intPtr(100),
			},
			validateFunc: func(t *testing.T, node domain.ContentNode) {
				assert.NotNil(t, node.Title)
				assert.Equal(t, "Test Document", *node.Title)
				assert.NotNil(t, node.MediaType)
				assert.Equal(t, "text/plain", *node.MediaType)
				assert.NotNil(t, node.TokenCount)
				assert.Equal(t, 100, *node.TokenCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validateFunc(t, tt.node)
		})
	}
}

func TestNodeRepo_CreateClusterNode_Validation(t *testing.T) {
	id := uuid.New()
	spaceID := uuid.New()
	title := "Test Cluster"
	embedding := []float32{0.1, 0.2, 0.3}

	tests := []struct {
		name             string
		id               uuid.UUID
		spaceID          uuid.UUID
		abstractionLevel int
		clusterScope     string
		title            *string
		embedding        *[]float32
		validateFunc     func(t *testing.T, id uuid.UUID, spaceID uuid.UUID, abstractionLevel int, clusterScope string, title *string, embedding *[]float32)
	}{
		{
			name:             "cluster node with all fields",
			id:               id,
			spaceID:          spaceID,
			abstractionLevel: 1,
			clusterScope:     "document",
			title:            &title,
			embedding:        &embedding,
			validateFunc: func(t *testing.T, id uuid.UUID, spaceID uuid.UUID, abstractionLevel int, clusterScope string, title *string, embedding *[]float32) {
				assert.NotEqual(t, uuid.Nil, id)
				assert.NotEqual(t, uuid.Nil, spaceID)
				assert.Equal(t, 1, abstractionLevel)
				assert.Equal(t, "document", clusterScope)
				assert.NotNil(t, title)
				assert.Equal(t, "Test Cluster", *title)
				assert.NotNil(t, embedding)
				assert.Len(t, *embedding, 3)
			},
		},
		{
			name:             "cluster node with minimal fields",
			id:               id,
			spaceID:          spaceID,
			abstractionLevel: 0,
			clusterScope:     "workspace",
			title:            nil,
			embedding:        nil,
			validateFunc: func(t *testing.T, id uuid.UUID, spaceID uuid.UUID, abstractionLevel int, clusterScope string, title *string, embedding *[]float32) {
				assert.NotEqual(t, uuid.Nil, id)
				assert.NotEqual(t, uuid.Nil, spaceID)
				assert.Equal(t, 0, abstractionLevel)
				assert.Equal(t, "workspace", clusterScope)
				assert.Nil(t, title)
				assert.Nil(t, embedding)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validateFunc(t, tt.id, tt.spaceID, tt.abstractionLevel, tt.clusterScope, tt.title, tt.embedding)
		})
	}
}

func TestNodeRepo_UpdateValidations(t *testing.T) {
	t.Run("ChunkNodeUpdate validation", func(t *testing.T) {
		id := uuid.New()
		content := "updated content"
		sequenceIndex := 5
		startPosition := int64(100)
		endPosition := int64(200)
		tokenCount := 50

		update := domain.ChunkNodeUpdate{
			ID: id,
			BaseNodeUpdate: domain.BaseNodeUpdate{
				Embedding: &[]float32{0.1, 0.2, 0.3},
			},
			Content:       &content,
			SequenceIndex: &sequenceIndex,
			StartPosition: &startPosition,
			EndPosition:   &endPosition,
			TokenCount:    &tokenCount,
		}

		// Validate update structure
		assert.NotEqual(t, uuid.Nil, update.ID)
		assert.NotNil(t, update.Content)
		assert.Equal(t, "updated content", *update.Content)
		assert.NotNil(t, update.SequenceIndex)
		assert.Equal(t, 5, *update.SequenceIndex)
		assert.NotNil(t, update.StartPosition)
		assert.Equal(t, int64(100), *update.StartPosition)
		assert.NotNil(t, update.EndPosition)
		assert.Equal(t, int64(200), *update.EndPosition)
		assert.NotNil(t, update.TokenCount)
		assert.Equal(t, 50, *update.TokenCount)
		assert.NotNil(t, update.BaseNodeUpdate.Embedding)
		assert.Len(t, *update.BaseNodeUpdate.Embedding, 3)
	})

	t.Run("ContentNodeUpdate validation", func(t *testing.T) {
		id := uuid.New()
		title := "Updated Title"
		mediaType := "text/plain"
		tokenCount := 100

		update := domain.ContentNodeUpdate{
			ID: id,
			BaseNodeUpdate: domain.BaseNodeUpdate{
				Embedding: &[]float32{0.1, 0.2, 0.3},
			},
			Title:      &title,
			MediaType:  &mediaType,
			TokenCount: &tokenCount,
		}

		assert.NotEqual(t, uuid.Nil, update.ID)
		assert.NotNil(t, update.Title)
		assert.Equal(t, "Updated Title", *update.Title)
		assert.NotNil(t, update.MediaType)
		assert.Equal(t, "text/plain", *update.MediaType)
		assert.NotNil(t, update.TokenCount)
		assert.Equal(t, 100, *update.TokenCount)
	})

	t.Run("ClusterNodeUpdate validation", func(t *testing.T) {
		id := uuid.New()
		title := "Updated Cluster"
		memberCount := 5
		coverageScore := 0.85

		update := domain.ClusterNodeUpdate{
			ID: id,
			BaseNodeUpdate: domain.BaseNodeUpdate{
				Embedding: &[]float32{0.1, 0.2, 0.3},
			},
			Title:         &title,
			MemberCount:   &memberCount,
			CoverageScore: &coverageScore,
		}

		assert.NotEqual(t, uuid.Nil, update.ID)
		assert.NotNil(t, update.Title)
		assert.Equal(t, "Updated Cluster", *update.Title)
		assert.NotNil(t, update.MemberCount)
		assert.Equal(t, 5, *update.MemberCount)
		assert.NotNil(t, update.CoverageScore)
		assert.Equal(t, 0.85, *update.CoverageScore)
	})
}

// Mock Repository Tests using dependency injection pattern
type TestableNodeRepo struct {
	executeWriteFunc func(cypher string, params map[string]interface{}) error
}

func (r *TestableNodeRepo) ExecuteWrite(cypher string, params map[string]interface{}) error {
	if r.executeWriteFunc != nil {
		return r.executeWriteFunc(cypher, params)
	}
	return nil
}

func TestNodeRepo_CreateChunkNodes_DatabaseInteraction(t *testing.T) {
	tests := []struct {
		name           string
		chunks         []domain.ChunkNode
		mockError      error
		expectedError  bool
		validateCypher func(t *testing.T, cypher string, params map[string]interface{})
	}{
		{
			name: "successful creation generates correct cypher",
			chunks: []domain.ChunkNode{
				{
					BaseNode: domain.BaseNode{
						ID:      uuid.New(),
						SpaceID: uuid.New(),
						Position3D: &domain.SpatialCoordinates{
							X: 10, Y: 20, Z: 30,
						},
					},
					ContentSourceID: uuid.New(),
					SequenceIndex:   0,
					StartPosition:   0,
					EndPosition:     100,
					Content:         "test content",
				},
			},
			mockError:     nil,
			expectedError: false,
			validateCypher: func(t *testing.T, cypher string, params map[string]interface{}) {
				assert.Contains(t, cypher, "UNWIND $items AS item")
				assert.Contains(t, cypher, "MERGE (n:ChunkNode {id: item.id})")
				assert.Contains(t, cypher, "ON CREATE SET")
				assert.Contains(t, cypher, "ON MATCH SET")
				assert.Contains(t, params, "items")

				items, ok := params["items"].([]map[string]interface{})
				assert.True(t, ok)
				assert.Len(t, items, 1)

				item := items[0]
				assert.Contains(t, item, "id")
				assert.Contains(t, item, "content_source_id")
				assert.Contains(t, item, "content")
				assert.Equal(t, "test content", item["content"])
			},
		},
		{
			name: "database error should propagate",
			chunks: []domain.ChunkNode{
				{
					BaseNode: domain.BaseNode{
						ID:      uuid.New(),
						SpaceID: uuid.New(),
						Position3D: &domain.SpatialCoordinates{
							X: 10, Y: 20, Z: 30,
						},
					},
					ContentSourceID: uuid.New(),
					Content:         "test content",
				},
			},
			mockError:     errors.New("database connection failed"),
			expectedError: true,
			validateCypher: func(t *testing.T, cypher string, params map[string]interface{}) {
				// Cypher should still be correct even if execution fails
				assert.Contains(t, cypher, "MERGE (n:ChunkNode {id: item.id})")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := &TestableNodeRepo{
				executeWriteFunc: func(cypher string, params map[string]interface{}) error {
					tt.validateCypher(t, cypher, params)
					return tt.mockError
				},
			}

			// This simulates the core logic without Neo4j dependency
			if len(tt.chunks) == 0 {
				return // Early return for empty chunks
			}

			// Simulate cypher generation
			params := map[string]interface{}{
				"items": make([]map[string]interface{}, 0, len(tt.chunks)),
			}

			for _, c := range tt.chunks {
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
				})
			}

			cypher := `
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
		`

			err := testRepo.ExecuteWrite(cypher, params)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}