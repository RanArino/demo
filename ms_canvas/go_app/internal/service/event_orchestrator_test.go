package service

import (
	"context"
	"testing"

	v1 "demo/ms_canvas/go_app/api/proto/public/v1"
	"demo/ms_canvas/go_app/internal/config"
	"demo/ms_canvas/go_app/internal/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNodeRepository is a mock implementation of NodeRepository
type MockNodeRepository struct {
	mock.Mock
}

func (m *MockNodeRepository) GetNodes(ctx context.Context, ids []string, filter *v1.NodeFilter) ([]*v1.Node, error) {
	args := m.Called(ctx, ids, filter)
	return args.Get(0).([]*v1.Node), args.Error(1)
}

func (m *MockNodeRepository) SearchNodes(ctx context.Context, filter *v1.NodeFilter, spatialBBox *v1.SpatialBoundingBox, limit int32) ([]*v1.Node, error) {
	args := m.Called(ctx, filter, spatialBBox, limit)
	return args.Get(0).([]*v1.Node), args.Error(1)
}

func (m *MockNodeRepository) CreateChunkNodes(ctx context.Context, chunks []*v1.ChunkNode) error {
	args := m.Called(ctx, chunks)
	return args.Error(0)
}

func (m *MockNodeRepository) CreateContentNodes(ctx context.Context, contents []*v1.ContentNode) error {
	args := m.Called(ctx, contents)
	return args.Error(0)
}

func (m *MockNodeRepository) CreateClusterNodes(ctx context.Context, clusters []*v1.ClusterNode) error {
	args := m.Called(ctx, clusters)
	return args.Error(0)
}

func (m *MockNodeRepository) UpdateClusterNode(ctx context.Context, update *v1.ClusterNode) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *MockNodeRepository) UpdateContentNode(ctx context.Context, update *v1.ContentNode) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *MockNodeRepository) UpdateChunkNode(ctx context.Context, update *v1.ChunkNode) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *MockNodeRepository) SoftDeleteNode(ctx context.Context, nodeID string) error {
	args := m.Called(ctx, nodeID)
	return args.Error(0)
}

// MockLinkRepository is a mock implementation of LinkRepository
type MockLinkRepository struct {
	mock.Mock
}

func (m *MockLinkRepository) CreateHierarchicalLinks(ctx context.Context, links []*v1.HierarchicalLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepository) CreateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepository) CreateStructuralLinks(ctx context.Context, links []*v1.StructuralLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepository) GetLinks(ctx context.Context, ids []string, filter *v1.BaseLinkFilter) ([]*v1.Link, error) {
	args := m.Called(ctx, ids, filter)
	return args.Get(0).([]*v1.Link), args.Error(1)
}

func (m *MockLinkRepository) GetLinksByNodes(ctx context.Context, nodeIDs []string, direction v1.Direction, query *v1.LinkQuery) ([]*v1.Link, error) {
	args := m.Called(ctx, nodeIDs, direction, query)
	return args.Get(0).([]*v1.Link), args.Error(1)
}

func (m *MockLinkRepository) UpdateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepository) UpdateStructuralLinks(ctx context.Context, links []*v1.StructuralLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepository) UpdateHierarchicalLinks(ctx context.Context, links []*v1.HierarchicalLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepository) DeleteLinks(ctx context.Context, linkIDs []string) error {
	args := m.Called(ctx, linkIDs)
	return args.Error(0)
}

func (m *MockLinkRepository) DeleteLinksForNodes(ctx context.Context, nodeIDs []string) error {
	args := m.Called(ctx, nodeIDs)
	return args.Error(0)
}

// MockPythonGateway is a mock implementation of the Python gateway
type MockPythonGateway struct {
	mock.Mock
}

func (m *MockPythonGateway) ChunkEmbed(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	// Return mock response for testing
	return &struct {
		Chunks     []*v1.ChunkNode `json:"chunks"`
		Embeddings []interface{}   `json:"embeddings"`
	}{}, args.Error(1)
}

func (m *MockPythonGateway) EmbedQuery(ctx context.Context, text string, config interface{}) (interface{}, error) {
	args := m.Called(ctx, text, config)
	// Return mock response for testing
	return &struct {
		Vector []float32 `json:"vector"`
	}{}, args.Error(1)
}

func (m *MockPythonGateway) Healthz(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	// Return mock response for testing
	return &struct {
		Status string `json:"status"`
	}{Status: "healthy"}, args.Error(1)
}

func (m *MockPythonGateway) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestEventOrchestrator_HandleDocumentProcessed(t *testing.T) {
	// Setup mocks
	mockNodeRepo := new(MockNodeRepository)
	mockLinkRepo := new(MockLinkRepository)

	// Create test configuration
	cfg := config.Config{
		Topics: config.Topics{
			DocumentProcessed: "document.processed",
		},
	}

	// Setup mock expectations
	mockNodeRepo.On("CreateContentNodes", mock.Anything, mock.Anything).Return(nil)

	// Create EventOrchestrator (python gateway not required for this test)
	orchestrator := NewEventOrchestrator(cfg, mockNodeRepo, mockLinkRepo, nil)

	// Create test event
	contentSourceID := uuid.New()
	event := events.DocumentProcessedEvent{
		ContentSourceID: contentSourceID,
		Status:          "PROCESSED",
	}

	// Execute
	err := orchestrator.HandleDocumentProcessed(event)

	// Assert
	assert.NoError(t, err)
	mockNodeRepo.AssertExpectations(t)
}

func TestEventOrchestrator_HandleDocumentProcessed_SkipsNonProcessed(t *testing.T) {
	// Setup
	mockNodeRepo := new(MockNodeRepository)
	mockLinkRepo := new(MockLinkRepository)

	cfg := config.Config{
		Topics: config.Topics{
			DocumentProcessed: "document.processed",
		},
	}

	orchestrator := NewEventOrchestrator(cfg, mockNodeRepo, mockLinkRepo, nil)

	// Create test event with non-PROCESSED status
	event := events.DocumentProcessedEvent{
		ContentSourceID: uuid.New(),
		Status:          "FAILED",
	}

	// Execute - should not call any repository methods
	err := orchestrator.HandleDocumentProcessed(event)

	// Assert
	assert.NoError(t, err)
	// For non-PROCESSED events, no methods should be called
	mockNodeRepo.AssertNotCalled(t, "CreateContentNodes")
}

func TestEventOrchestrator_HandleDocumentProcessed_WithBlobHash(t *testing.T) {
	// Setup
	mockNodeRepo := new(MockNodeRepository)
	mockLinkRepo := new(MockLinkRepository)

	cfg := config.Config{
		Topics: config.Topics{
			DocumentProcessed: "document.processed",
		},
	}

	mockNodeRepo.On("CreateContentNodes", mock.Anything, mock.Anything).Return(nil)

	orchestrator := NewEventOrchestrator(cfg, mockNodeRepo, mockLinkRepo, nil)

	// Create test event with blob hash
	contentSourceID := uuid.New()
	blobHash := "abc123"
	event := events.DocumentProcessedEvent{
		ContentSourceID:   contentSourceID,
		ProcessedBlobHash: &blobHash,
		Status:            "PROCESSED",
	}

	// Execute
	err := orchestrator.HandleDocumentProcessed(event)

	// Assert
	assert.NoError(t, err)
	mockNodeRepo.AssertExpectations(t)
}

func TestEventOrchestrator_HandleDocumentProcessed_CreateContentNodeError(t *testing.T) {
	// Setup
	mockNodeRepo := new(MockNodeRepository)
	mockLinkRepo := new(MockLinkRepository)

	cfg := config.Config{
		Topics: config.Topics{
			DocumentProcessed: "document.processed",
		},
	}

	mockNodeRepo.On("CreateContentNodes", mock.Anything, mock.Anything).Return(assert.AnError)

	orchestrator := NewEventOrchestrator(cfg, mockNodeRepo, mockLinkRepo, nil)

	// Create test event
	event := events.DocumentProcessedEvent{
		ContentSourceID: uuid.New(),
		Status:          "PROCESSED",
	}

	// Execute
	err := orchestrator.HandleDocumentProcessed(event)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create content node")
	mockNodeRepo.AssertExpectations(t)
}

// Additional task-related tests are implemented in the task executor test suite.
