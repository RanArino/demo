package service

import (
	"context"
	"errors"
	"testing"

	v1 "demo/ms_canvas/go_app/api/proto/public/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock repositories
type MockNodeRepo struct {
	mock.Mock
}

func (m *MockNodeRepo) GetNodes(ctx context.Context, ids []string, filter *v1.NodeFilter) ([]*v1.Node, error) {
	args := m.Called(ctx, ids, filter)
	return args.Get(0).([]*v1.Node), args.Error(1)
}

func (m *MockNodeRepo) SearchNodes(ctx context.Context, filter *v1.NodeFilter, spatialBBox *v1.SpatialBoundingBox, limit int32) ([]*v1.Node, error) {
	args := m.Called(ctx, filter, spatialBBox, limit)
	return args.Get(0).([]*v1.Node), args.Error(1)
}

func (m *MockNodeRepo) CreateChunkNodes(ctx context.Context, chunks []*v1.ChunkNode) error {
	args := m.Called(ctx, chunks)
	return args.Error(0)
}

func (m *MockNodeRepo) CreateContentNodes(ctx context.Context, contents []*v1.ContentNode) error {
	args := m.Called(ctx, contents)
	return args.Error(0)
}

func (m *MockNodeRepo) CreateClusterNodes(ctx context.Context, clusters []*v1.ClusterNode) error {
	args := m.Called(ctx, clusters)
	return args.Error(0)
}

func (m *MockNodeRepo) UpdateClusterNode(ctx context.Context, update *v1.ClusterNode) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *MockNodeRepo) UpdateContentNode(ctx context.Context, update *v1.ContentNode) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *MockNodeRepo) UpdateChunkNode(ctx context.Context, update *v1.ChunkNode) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *MockNodeRepo) SoftDeleteNodes(ctx context.Context, nodeIDs []string) error {
	args := m.Called(ctx, nodeIDs)
	return args.Error(0)
}

type MockLinkRepoForNodeService struct {
	mock.Mock
}

func (m *MockLinkRepoForNodeService) CreateHierarchicalLinks(ctx context.Context, links []*v1.HierarchicalLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepoForNodeService) CreateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepoForNodeService) CreateStructuralLinks(ctx context.Context, links []*v1.StructuralLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepoForNodeService) GetLinks(ctx context.Context, ids []string, query *v1.LinkQuery) ([]*v1.Link, error) {
	args := m.Called(ctx, ids, query)
	return args.Get(0).([]*v1.Link), args.Error(1)
}

func (m *MockLinkRepoForNodeService) GetLinksByNodes(ctx context.Context, nodeIDs []string, direction v1.Direction, query *v1.LinkQuery) ([]*v1.Link, error) {
	args := m.Called(ctx, nodeIDs, direction, query)
	return args.Get(0).([]*v1.Link), args.Error(1)
}

func (m *MockLinkRepoForNodeService) UpdateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepoForNodeService) UpdateStructuralLinks(ctx context.Context, links []*v1.StructuralLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepoForNodeService) UpdateHierarchicalLinks(ctx context.Context, links []*v1.HierarchicalLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepoForNodeService) DeleteLinks(ctx context.Context, linkIDs []string) error {
	args := m.Called(ctx, linkIDs)
	return args.Error(0)
}

func (m *MockLinkRepoForNodeService) DeleteLinksForNodes(ctx context.Context, nodeIDs []string) error {
	args := m.Called(ctx, nodeIDs)
	return args.Error(0)
}

// Helper functions to create test data
func createTestContentNode(id string) *v1.Node {
	return &v1.Node{
		Node: &v1.Node_Content{
			Content: &v1.ContentNode{
				Base: &v1.BaseNode{
					Id:               id,
					SpaceId:          "test-space",
					AbstractionLevel: 1,
					ContextType:      "document",
					Keywords:         []string{"test", "content"},
					DisplayContent:   stringPtr("Test Content"),
					SemanticDensity:  float64Ptr(0.8),
					Position_3D: &v1.SpatialCoordinates{
						X: 10,
						Y: 20,
						Z: 30,
					},
					IsPositionLocked: boolPtr(false),
					Visibility:       boolPtr(true),
					DisplayProps: &v1.DisplayProps{
						Size:    100,
						Opacity: 0.9,
						Shape:   "circle",
						Color:   "#ff0000",
					},
					EngagementScore: &v1.EngagementScore{
						CanvasScore:  0.7,
						ChatScore:    0.6,
						OverallScore: 0.65,
					},
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
				ContentSourceId: "source-123",
				Title:           stringPtr("Test Title"),
				MediaType:       stringPtr("text"),
				Source:          stringPtr("test-source"),
				TokenCount:      int32Ptr(100),
			},
		},
	}
}

func createTestChunkNode(id string) *v1.Node {
	return &v1.Node{
		Node: &v1.Node_Chunk{
			Chunk: &v1.ChunkNode{
				Base: &v1.BaseNode{
					Id:               id,
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
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
				ContentSourceId: "source-456",
				SequenceIndex:   1,
				ChunkType:       "paragraph",
				StartPosition:   int64Ptr(0),
				EndPosition:     int64Ptr(100),
				Content:         "Test chunk content",
				TokenCount:      int32Ptr(50),
			},
		},
	}
}

func createTestClusterNode(id string) *v1.Node {
	return &v1.Node{
		Node: &v1.Node_Cluster{
			Cluster: &v1.ClusterNode{
				Base: &v1.BaseNode{
					Id:               id,
					SpaceId:          "test-space",
					AbstractionLevel: 0,
					ContextType:      "cluster",
					Keywords:         []string{"test", "cluster"},
					DisplayContent:   stringPtr("Test Cluster"),
					SemanticDensity:  float64Ptr(0.6),
					Position_3D: &v1.SpatialCoordinates{
						X: 5,
						Y: 15,
						Z: 25,
					},
					IsPositionLocked: boolPtr(false),
					Visibility:       boolPtr(true),
					DisplayProps: &v1.DisplayProps{
						Size:    200,
						Opacity: 0.7,
						Shape:   "hexagon",
						Color:   "#0000ff",
					},
					EngagementScore: &v1.EngagementScore{
						CanvasScore:  0.9,
						ChatScore:    0.8,
						OverallScore: 0.85,
					},
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
				ClusterScope:  "document",
				Title:         stringPtr("Test Cluster"),
				MemberCount:   int32Ptr(10),
				CoverageScore: float64Ptr(0.95),
			},
		},
	}
}

func createTestLinks() []*v1.Link {
	nodeID1 := uuid.New().String()
	nodeID2 := uuid.New().String()
	nodeID3 := uuid.New().String()

	link1 := &v1.Link{
		Link: &v1.Link_Structural{
			Structural: &v1.StructuralLink{
				Base: &v1.BaseLink{
					SourceId:  nodeID1,
					TargetId:  nodeID2,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
				ConnectionType:  v1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_EVIDENCE_BASED,
				ConfidenceScore: 0.8,
				Description:     stringPtr("Test link"),
			},
		},
	}

	link2 := &v1.Link{
		Link: &v1.Link_Semantic{
			Semantic: &v1.SemanticLink{
				Base: &v1.BaseLink{
					SourceId:  nodeID2,
					TargetId:  nodeID3,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
				ConnectionType: v1.SemanticConnectionType_SEMANTIC_CONNECTION_TYPE_INTRA_LEVEL_INTRA_PARENT,
			},
		},
	}

	return []*v1.Link{link1, link2}
}

func TestNodeService_GetNodes(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - GetNodes with valid IDs", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodeID1 := uuid.New().String()
		nodeID2 := uuid.New().String()
		node1 := createTestContentNode(nodeID1)
		node2 := createTestChunkNode(nodeID2)

		// Setup expectations
		mockNodeRepo.On("GetNodes", ctx, []string{nodeID1, nodeID2}, (*v1.NodeFilter)(nil)).Return([]*v1.Node{node1, node2}, nil)

		// Execute
		result, err := service.GetNodes(ctx, []string{nodeID1, nodeID2})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, nodeID1, getNodeID(result[0]))
		assert.Equal(t, nodeID2, getNodeID(result[1]))
		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Success - GetNodes with invalid ID", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		invalidID := "invalid-uuid"

		// Setup expectations
		mockNodeRepo.On("GetNodes", ctx, []string{invalidID}, (*v1.NodeFilter)(nil)).Return([]*v1.Node{}, nil)

		// Execute
		result, err := service.GetNodes(ctx, []string{invalidID})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Error - Repository error", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodeID1 := uuid.New().String()
		repoError := errors.New("database connection failed")

		// Setup expectations
		mockNodeRepo.On("GetNodes", ctx, []string{nodeID1}, (*v1.NodeFilter)(nil)).Return([]*v1.Node{}, repoError)

		// Execute
		result, err := service.GetNodes(ctx, []string{nodeID1})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get nodes")
		mockNodeRepo.AssertExpectations(t)
	})
}

func TestNodeService_GetNeighbors(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - GetNeighbors with links", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		links := createTestLinks()
		// Extract node IDs from the test links
		nodeID1 := links[0].GetStructural().Base.SourceId
		nodeID2 := links[0].GetStructural().Base.TargetId
		nodeID3 := links[1].GetSemantic().Base.TargetId
		neighborNode := createTestChunkNode(nodeID2)

		// Setup expectations
		mockLinkRepo.On("GetLinksByNodes", ctx, []string{nodeID1}, v1.Direction_DIRECTION_OUTGOING, (*v1.LinkQuery)(nil)).Return(links, nil)
		mockNodeRepo.On("GetNodes", ctx, mock.MatchedBy(func(ids []string) bool {
			return len(ids) == 2 &&
				((ids[0] == nodeID2 && ids[1] == nodeID3) || (ids[0] == nodeID3 && ids[1] == nodeID2))
		}), (*v1.NodeFilter)(nil)).Return([]*v1.Node{neighborNode}, nil)

		// Execute
		result, err := service.GetNeighbors(ctx, []string{nodeID1}, v1.Direction_DIRECTION_OUTGOING, nil, 10, true)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Contains(t, result, nodeID1)

		neighbors := result[nodeID1].Neighbors
		assert.Len(t, neighbors, 1)
		assert.Equal(t, nodeID2, getNodeID(neighbors[0].Node))
		assert.Equal(t, v1.LinkType_LINK_TYPE_STRUCTURAL, neighbors[0].LinkType)

		mockLinkRepo.AssertExpectations(t)
		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Success - GetNeighbors with no links", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodeID1 := uuid.New().String()

		// Setup expectations
		mockLinkRepo.On("GetLinksByNodes", ctx, []string{nodeID1}, v1.Direction_DIRECTION_OUTGOING, (*v1.LinkQuery)(nil)).Return([]*v1.Link{}, nil)
		mockNodeRepo.On("GetNodes", ctx, []string{}, (*v1.NodeFilter)(nil)).Return([]*v1.Node{}, nil)

		// Execute
		result, err := service.GetNeighbors(ctx, []string{nodeID1}, v1.Direction_DIRECTION_OUTGOING, nil, 10, true)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Contains(t, result, nodeID1)

		neighbors := result[nodeID1].Neighbors
		assert.Len(t, neighbors, 0)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Error - Link repository error", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodeID1 := uuid.New().String()
		repoError := errors.New("link repository error")

		// Setup expectations
		mockLinkRepo.On("GetLinksByNodes", ctx, []string{nodeID1}, v1.Direction_DIRECTION_OUTGOING, (*v1.LinkQuery)(nil)).Return([]*v1.Link{}, repoError)

		// Execute
		result, err := service.GetNeighbors(ctx, []string{nodeID1}, v1.Direction_DIRECTION_OUTGOING, nil, 10, true)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get links by nodes")

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Error - Node repository error", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		links := createTestLinks()
		// Extract node IDs from the test links
		nodeID1 := links[0].GetStructural().Base.SourceId
		nodeID2 := links[0].GetStructural().Base.TargetId
		nodeID3 := links[1].GetSemantic().Base.TargetId
		repoError := errors.New("node repository error")

		// Setup expectations
		mockLinkRepo.On("GetLinksByNodes", ctx, []string{nodeID1}, v1.Direction_DIRECTION_OUTGOING, (*v1.LinkQuery)(nil)).Return(links, nil)
		mockNodeRepo.On("GetNodes", ctx, mock.MatchedBy(func(ids []string) bool {
			return len(ids) == 2 &&
				((ids[0] == nodeID2 && ids[1] == nodeID3) || (ids[0] == nodeID3 && ids[1] == nodeID2))
		}), (*v1.NodeFilter)(nil)).Return([]*v1.Node{}, repoError)

		// Execute
		result, err := service.GetNeighbors(ctx, []string{nodeID1}, v1.Direction_DIRECTION_OUTGOING, nil, 10, true)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get neighbor nodes")

		mockLinkRepo.AssertExpectations(t)
		mockNodeRepo.AssertExpectations(t)
	})
}

func TestNodeService_UpdateNodes(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - UpdateNodes with mixed types", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodeID1 := uuid.New().String()
		nodeID2 := uuid.New().String()
		contentNode := createTestContentNode(nodeID1)
		chunkNode := createTestChunkNode(nodeID2)
		nodes := []*v1.Node{contentNode, chunkNode}

		// Setup expectations
		mockNodeRepo.On("UpdateContentNode", ctx, mock.MatchedBy(func(node *v1.ContentNode) bool {
			return node.Base.Id == nodeID1
		})).Return(nil)
		mockNodeRepo.On("UpdateChunkNode", ctx, mock.MatchedBy(func(node *v1.ChunkNode) bool {
			return node.Base.Id == nodeID2
		})).Return(nil)

		// Execute
		result, err := service.UpdateNodes(ctx, nodes)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, nodeID1, getNodeID(result[0]))
		assert.Equal(t, nodeID2, getNodeID(result[1]))

		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Success - UpdateNodes with content nodes only", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodeID1 := uuid.New().String()
		contentNode := createTestContentNode(nodeID1)
		nodes := []*v1.Node{contentNode}

		// Setup expectations
		mockNodeRepo.On("UpdateContentNode", ctx, mock.MatchedBy(func(node *v1.ContentNode) bool {
			return node.Base.Id == nodeID1
		})).Return(nil)

		// Execute
		result, err := service.UpdateNodes(ctx, nodes)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, nodeID1, getNodeID(result[0]))

		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Success - UpdateNodes with cluster nodes only", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		clusterNodeID := uuid.New().String()
		clusterNode := createTestClusterNode(clusterNodeID)
		nodes := []*v1.Node{clusterNode}

		// Setup expectations
		mockNodeRepo.On("UpdateClusterNode", ctx, mock.MatchedBy(func(node *v1.ClusterNode) bool {
			return node.Base.Id == clusterNodeID
		})).Return(nil)

		// Execute
		result, err := service.UpdateNodes(ctx, nodes)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Error - Repository error on content node update", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodeID1 := uuid.New().String()
		contentNode := createTestContentNode(nodeID1)
		nodes := []*v1.Node{contentNode}

		repoError := errors.New("content node update failed")

		// Setup expectations
		mockNodeRepo.On("UpdateContentNode", ctx, mock.MatchedBy(func(node *v1.ContentNode) bool {
			return node.Base.Id == nodeID1
		})).Return(repoError)

		// Execute
		result, err := service.UpdateNodes(ctx, nodes)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "multiple update errors")

		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Error - Repository error on chunk node update", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodeID2 := uuid.New().String()
		chunkNode := createTestChunkNode(nodeID2)
		nodes := []*v1.Node{chunkNode}

		repoError := errors.New("chunk node update failed")

		// Setup expectations
		mockNodeRepo.On("UpdateChunkNode", ctx, mock.MatchedBy(func(node *v1.ChunkNode) bool {
			return node.Base.Id == nodeID2
		})).Return(repoError)

		// Execute
		result, err := service.UpdateNodes(ctx, nodes)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "multiple update errors")

		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Success - UpdateNodes with empty list", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)
		nodes := []*v1.Node{}

		// Execute
		result, err := service.UpdateNodes(ctx, nodes)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 0)

		// No repository calls should be made
		mockNodeRepo.AssertNotCalled(t, "UpdateContentNode")
		mockNodeRepo.AssertNotCalled(t, "UpdateChunkNode")
		mockNodeRepo.AssertNotCalled(t, "UpdateClusterNode")
	})
}

func TestNodeService_SearchNodes(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - SearchNodes", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)

		filter := &v1.NodeFilter{
			SpaceId:  stringPtr("test-space"),
			Keywords: []string{"test"},
		}
		spatialBBox := &v1.SpatialBoundingBox{
			MinCoords: &v1.SpatialCoordinates{
				X: 0,
				Y: 0,
				Z: 0,
			},
			MaxCoords: &v1.SpatialCoordinates{
				X: 10,
				Y: 10,
				Z: 10,
			},
		}

		// Create test nodes
		searchResult := []*v1.Node{
			createTestContentNode(uuid.New().String()),
			createTestChunkNode(uuid.New().String()),
		}

		// Setup expectations
		mockNodeRepo.On("SearchNodes", ctx, filter, spatialBBox, int32(10)).Return(searchResult, nil)

		// Execute
		result, err := service.SearchNodes(ctx, filter, spatialBBox, 10)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, searchResult, result)

		mockNodeRepo.AssertExpectations(t)
	})

	t.Run("Error - Repository error", func(t *testing.T) {
		mockNodeRepo := &MockNodeRepo{}
		mockLinkRepo := &MockLinkRepoForNodeService{}
		service := NewNodeService(mockNodeRepo, mockLinkRepo)

		filter := &v1.NodeFilter{
			SpaceId:  stringPtr("test-space"),
			Keywords: []string{"test"},
		}
		spatialBBox := &v1.SpatialBoundingBox{
			MinCoords: &v1.SpatialCoordinates{
				X: 0,
				Y: 0,
				Z: 0,
			},
			MaxCoords: &v1.SpatialCoordinates{
				X: 10,
				Y: 10,
				Z: 10,
			},
		}
		repoError := errors.New("search failed")

		// Setup expectations
		mockNodeRepo.On("SearchNodes", ctx, filter, spatialBBox, int32(10)).Return([]*v1.Node{}, repoError)

		// Execute
		result, err := service.SearchNodes(ctx, filter, spatialBBox, 10)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to search nodes")

		mockNodeRepo.AssertExpectations(t)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getNodeID with content node", func(t *testing.T) {
		nodeID := uuid.New().String()
		node := createTestContentNode(nodeID)

		result := getNodeID(node)
		assert.Equal(t, nodeID, result)
	})

	t.Run("getNodeID with chunk node", func(t *testing.T) {
		nodeID := uuid.New().String()
		node := createTestChunkNode(nodeID)

		result := getNodeID(node)
		assert.Equal(t, nodeID, result)
	})

	t.Run("getNodeID with cluster node", func(t *testing.T) {
		nodeID := uuid.New().String()
		node := createTestClusterNode(nodeID)

		result := getNodeID(node)
		assert.Equal(t, nodeID, result)
	})

	t.Run("getNodeID with nil node", func(t *testing.T) {
		result := getNodeID(nil)
		assert.Equal(t, "nil", result)
	})

	t.Run("linkTypeToString with hierarchical", func(t *testing.T) {
		result := linkTypeToString(v1.LinkType_LINK_TYPE_HIERARCHICAL)
		assert.Equal(t, "hierarchical", result)
	})

	t.Run("linkTypeToString with semantic", func(t *testing.T) {
		result := linkTypeToString(v1.LinkType_LINK_TYPE_SEMANTIC)
		assert.Equal(t, "semantic", result)
	})

	t.Run("linkTypeToString with structural", func(t *testing.T) {
		result := linkTypeToString(v1.LinkType_LINK_TYPE_STRUCTURAL)
		assert.Equal(t, "structural", result)
	})

	t.Run("linkTypeToString with unspecified", func(t *testing.T) {
		result := linkTypeToString(v1.LinkType_LINK_TYPE_UNSPECIFIED)
		assert.Equal(t, "", result)
	})
}
