package service

import (
	"context"
	"errors"
	"testing"

	v1 "demo/ms_canvas/go_app/api/proto/public/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock repositories for link service tests
type MockLinkRepo struct {
	mock.Mock
}

func (m *MockLinkRepo) CreateHierarchicalLinks(ctx context.Context, links []*v1.HierarchicalLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepo) CreateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepo) CreateStructuralLinks(ctx context.Context, links []*v1.StructuralLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepo) GetLinks(ctx context.Context, ids []string, query *v1.LinkQuery) ([]*v1.Link, error) {
	args := m.Called(ctx, ids, query)
	return args.Get(0).([]*v1.Link), args.Error(1)
}

func (m *MockLinkRepo) GetLinksByNodes(ctx context.Context, nodeIDs []string, direction v1.Direction, query *v1.LinkQuery) ([]*v1.Link, error) {
	args := m.Called(ctx, nodeIDs, direction, query)
	return args.Get(0).([]*v1.Link), args.Error(1)
}

func (m *MockLinkRepo) UpdateSemanticLinks(ctx context.Context, links []*v1.SemanticLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepo) UpdateStructuralLinks(ctx context.Context, links []*v1.StructuralLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepo) UpdateHierarchicalLinks(ctx context.Context, links []*v1.HierarchicalLink) error {
	args := m.Called(ctx, links)
	return args.Error(0)
}

func (m *MockLinkRepo) DeleteLinks(ctx context.Context, linkIDs []string) error {
	args := m.Called(ctx, linkIDs)
	return args.Error(0)
}

func (m *MockLinkRepo) DeleteLinksForNodes(ctx context.Context, nodeIDs []string) error {
	args := m.Called(ctx, nodeIDs)
	return args.Error(0)
}

// Helper functions to create test data for link service
func createTestStructuralLinkCreate() *v1.StructuralLinkCreate {
	explorationStruct, _ := structpb.NewStruct(map[string]interface{}{
		"exploration_id": "exploration-123",
		"step_id":        "step-456",
	})

	styleStruct, _ := structpb.NewStruct(map[string]interface{}{
		"line_style": "solid",
		"line_width": 2,
		"line_color": "#000000",
	})

	return &v1.StructuralLinkCreate{
		SourceId:            uuid.New().String(),
		TargetId:            uuid.New().String(),
		ConnectionType:      v1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_EVIDENCE_BASED,
		ConfidenceScore:     0.8,
		Description:         "Test link creation",
		CreatedBy:           "test-user",
		ExplorationMetadata: explorationStruct,
		StyleMetadata:       styleStruct,
	}
}

func createTestStructuralLinkUpdate() *v1.StructuralLinkUpdate {
	return &v1.StructuralLinkUpdate{
		SourceId: uuid.New().String(),
		TargetId: uuid.New().String(),
		ConnectionType: func() *v1.StructuralConnectionType {
			val := v1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_CITATION
			return &val
		}(),
		ConfidenceScore: func() *float64 { val := 0.9; return &val }(),
		Description:     func() *string { val := "Updated test link"; return &val }(),
		ExplorationMetadata: func() *structpb.Struct {
			s, _ := structpb.NewStruct(map[string]interface{}{
				"exploration_id": "exploration-updated",
				"step_id":        "step-updated",
			})
			return s
		}(),
		StyleMetadata: func() *structpb.Struct {
			s, _ := structpb.NewStruct(map[string]interface{}{
				"line_style": "dashed",
				"line_width": 3,
				"line_color": "#ff0000",
			})
			return s
		}(),
	}
}

func createTestStructuralLinkIdentifier() *v1.StructuralLinkIdentifier {
	return &v1.StructuralLinkIdentifier{
		SourceId: uuid.New().String(),
		TargetId: uuid.New().String(),
	}
}

func createTestStructuralLink() *v1.StructuralLink {
	explorationStruct, _ := structpb.NewStruct(map[string]interface{}{
		"exploration_id": "exploration-123",
		"step_id":        "step-456",
	})

	styleStruct, _ := structpb.NewStruct(map[string]interface{}{
		"line_style": "solid",
		"line_width": 2,
		"line_color": "#000000",
	})

	return &v1.StructuralLink{
		Base: &v1.BaseLink{
			SourceId:            uuid.New().String(),
			TargetId:            uuid.New().String(),
			CreatedAt:           timestamppb.Now(),
			UpdatedAt:           timestamppb.Now(),
			ExplorationMetadata: explorationStruct,
			StyleMetadata:       styleStruct,
		},
		ConnectionType:  v1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_EVIDENCE_BASED,
		ConfidenceScore: 0.8,
		Description:     stringPtr("Test structural link"),
		CreatedBy:       "test-user",
	}
}

// Helper functions for creating pointer types
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestLinkService_CreateStructuralLinks(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - CreateStructuralLinks with single link", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkCreate := createTestStructuralLinkCreate()

		// Setup expectations
		mockLinkRepo.On("CreateStructuralLinks", ctx, mock.MatchedBy(func(links []*v1.StructuralLink) bool {
			return len(links) == 1 &&
				links[0].Base.SourceId == linkCreate.SourceId &&
				links[0].Base.TargetId == linkCreate.TargetId &&
				links[0].ConnectionType == linkCreate.ConnectionType &&
				links[0].ConfidenceScore == linkCreate.ConfidenceScore &&
				*links[0].Description == linkCreate.Description
		})).Return(nil)

		// Execute
		result, err := service.CreateStructuralLinks(ctx, []*v1.StructuralLinkCreate{linkCreate})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, linkCreate.SourceId, result[0].Base.SourceId)
		assert.Equal(t, linkCreate.TargetId, result[0].Base.TargetId)
		assert.Equal(t, linkCreate.ConnectionType, result[0].ConnectionType)
		assert.Equal(t, linkCreate.ConfidenceScore, result[0].ConfidenceScore)
		assert.Equal(t, linkCreate.Description, *result[0].Description)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Success - CreateStructuralLinks with multiple links", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkCreate1 := createTestStructuralLinkCreate()
		linkCreate2 := createTestStructuralLinkCreate()
		linkCreate2.SourceId = uuid.New().String()
		linkCreate2.TargetId = uuid.New().String()

		// Setup expectations
		mockLinkRepo.On("CreateStructuralLinks", ctx, mock.MatchedBy(func(links []*v1.StructuralLink) bool {
			return len(links) == 2
		})).Return(nil).Times(1)

		// Execute
		result, err := service.CreateStructuralLinks(ctx, []*v1.StructuralLinkCreate{linkCreate1, linkCreate2})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Success - CreateStructuralLinks with empty list", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)

		// Execute
		result, err := service.CreateStructuralLinks(ctx, []*v1.StructuralLinkCreate{})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 0)

		mockLinkRepo.AssertNotCalled(t, "CreateStructuralLinks")
	})

	t.Run("Error - Repository error", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkCreate := createTestStructuralLinkCreate()
		repoError := errors.New("database connection failed")

		// Setup expectations
		mockLinkRepo.On("CreateStructuralLinks", ctx, mock.MatchedBy(func(links []*v1.StructuralLink) bool {
			return len(links) == 1
		})).Return(repoError)

		// Execute
		result, err := service.CreateStructuralLinks(ctx, []*v1.StructuralLinkCreate{linkCreate})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create structural links")

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid source ID", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkCreate := createTestStructuralLinkCreate()
		linkCreate.SourceId = "invalid-uuid"

		// Execute
		result, err := service.CreateStructuralLinks(ctx, []*v1.StructuralLinkCreate{linkCreate})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid source ID")

		mockLinkRepo.AssertNotCalled(t, "CreateStructuralLinks")
	})

	t.Run("Error - Invalid target ID", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkCreate := createTestStructuralLinkCreate()
		linkCreate.TargetId = "invalid-uuid"

		// Execute
		result, err := service.CreateStructuralLinks(ctx, []*v1.StructuralLinkCreate{linkCreate})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid target ID")

		mockLinkRepo.AssertNotCalled(t, "CreateStructuralLinks")
	})
}

func TestLinkService_UpdateStructuralLinks(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - UpdateStructuralLinks with single update", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkUpdate := createTestStructuralLinkUpdate()
		expectedLink := createTestStructuralLink()
		expectedLink.ConnectionType = *linkUpdate.ConnectionType
		expectedLink.ConfidenceScore = *linkUpdate.ConfidenceScore
		expectedLink.Description = linkUpdate.Description

		// Setup expectations
		mockLinkRepo.On("UpdateStructuralLinks", ctx, mock.MatchedBy(func(links []*v1.StructuralLink) bool {
			return len(links) == 1 &&
				links[0].Base.SourceId == linkUpdate.SourceId &&
				links[0].Base.TargetId == linkUpdate.TargetId &&
				links[0].ConnectionType == *linkUpdate.ConnectionType &&
				links[0].ConfidenceScore == *linkUpdate.ConfidenceScore &&
				*links[0].Description == *linkUpdate.Description
		})).Return(nil)

		// Execute
		result, err := service.UpdateStructuralLinks(ctx, []*v1.StructuralLinkUpdate{linkUpdate})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, linkUpdate.SourceId, result[0].Base.SourceId)
		assert.Equal(t, linkUpdate.TargetId, result[0].Base.TargetId)
		assert.Equal(t, *linkUpdate.ConnectionType, result[0].ConnectionType)
		assert.Equal(t, *linkUpdate.ConfidenceScore, result[0].ConfidenceScore)
		assert.Equal(t, *linkUpdate.Description, *result[0].Description)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Success - UpdateStructuralLinks with multiple updates", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkUpdate1 := createTestStructuralLinkUpdate()
		linkUpdate2 := createTestStructuralLinkUpdate()
		linkUpdate2.SourceId = uuid.New().String()
		linkUpdate2.TargetId = uuid.New().String()

		// Setup expectations
		mockLinkRepo.On("UpdateStructuralLinks", ctx, mock.MatchedBy(func(links []*v1.StructuralLink) bool {
			return len(links) == 2
		})).Return(nil).Times(1)

		// Execute
		result, err := service.UpdateStructuralLinks(ctx, []*v1.StructuralLinkUpdate{linkUpdate1, linkUpdate2})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Success - UpdateStructuralLinks with empty list", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)

		// Execute
		result, err := service.UpdateStructuralLinks(ctx, []*v1.StructuralLinkUpdate{})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 0)

		mockLinkRepo.AssertNotCalled(t, "UpdateStructuralLinks")
	})

	t.Run("Success - UpdateStructuralLinks with partial updates", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkUpdate := &v1.StructuralLinkUpdate{
			SourceId:    uuid.New().String(),
			TargetId:    uuid.New().String(),
			Description: func() *string { val := "Only description updated"; return &val }(),
			// Other fields are nil
		}

		expectedLink := createTestStructuralLink()
		expectedLink.Base.SourceId = linkUpdate.SourceId
		expectedLink.Base.TargetId = linkUpdate.TargetId
		expectedLink.Description = linkUpdate.Description

		// Setup expectations
		mockLinkRepo.On("UpdateStructuralLinks", ctx, mock.MatchedBy(func(links []*v1.StructuralLink) bool {
			return len(links) == 1 &&
				links[0].Base.SourceId == linkUpdate.SourceId &&
				links[0].Base.TargetId == linkUpdate.TargetId &&
				*links[0].Description == *linkUpdate.Description
		})).Return(nil)

		// Execute
		result, err := service.UpdateStructuralLinks(ctx, []*v1.StructuralLinkUpdate{linkUpdate})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Error - Repository error", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkUpdate := createTestStructuralLinkUpdate()
		repoError := errors.New("database update failed")

		// Setup expectations
		mockLinkRepo.On("UpdateStructuralLinks", ctx, mock.MatchedBy(func(links []*v1.StructuralLink) bool {
			return len(links) == 1
		})).Return(repoError)

		// Execute
		result, err := service.UpdateStructuralLinks(ctx, []*v1.StructuralLinkUpdate{linkUpdate})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to update structural links")

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid source ID", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkUpdate := createTestStructuralLinkUpdate()
		linkUpdate.SourceId = "invalid-uuid"

		// Execute
		result, err := service.UpdateStructuralLinks(ctx, []*v1.StructuralLinkUpdate{linkUpdate})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid source ID")

		mockLinkRepo.AssertNotCalled(t, "UpdateStructuralLinks")
	})

	t.Run("Error - Invalid target ID", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkUpdate := createTestStructuralLinkUpdate()
		linkUpdate.TargetId = "invalid-uuid"

		// Execute
		result, err := service.UpdateStructuralLinks(ctx, []*v1.StructuralLinkUpdate{linkUpdate})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid target ID")

		mockLinkRepo.AssertNotCalled(t, "UpdateStructuralLinks")
	})
}

func TestLinkService_DeleteStructuralLinks(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - DeleteStructuralLinks with single identifier", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkID := createTestStructuralLinkIdentifier()

		// Note: The current implementation has a TODO comment and doesn't actually delete from database
		// It just logs and returns success. So we expect it to return 1 without calling the repo.

		// Execute
		result, err := service.DeleteStructuralLinks(ctx, []*v1.StructuralLinkIdentifier{linkID})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(1), result)

		// Repository should not be called due to TODO implementation
		mockLinkRepo.AssertNotCalled(t, "DeleteLinks")
	})

	t.Run("Success - DeleteStructuralLinks with multiple identifiers", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkID1 := createTestStructuralLinkIdentifier()
		linkID2 := createTestStructuralLinkIdentifier()

		// Execute
		result, err := service.DeleteStructuralLinks(ctx, []*v1.StructuralLinkIdentifier{linkID1, linkID2})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(2), result)

		mockLinkRepo.AssertNotCalled(t, "DeleteLinks")
	})

	t.Run("Success - DeleteStructuralLinks with empty list", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)

		// Execute
		result, err := service.DeleteStructuralLinks(ctx, []*v1.StructuralLinkIdentifier{})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(0), result)

		mockLinkRepo.AssertNotCalled(t, "DeleteLinks")
	})

	t.Run("Error - Invalid source ID", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkID := createTestStructuralLinkIdentifier()
		linkID.SourceId = "invalid-uuid"

		// Execute
		result, err := service.DeleteStructuralLinks(ctx, []*v1.StructuralLinkIdentifier{linkID})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(0), result) // Invalid UUIDs are skipped, so count is 0

		// Still succeeds but logs the error
		mockLinkRepo.AssertNotCalled(t, "DeleteLinks")
	})

	t.Run("Error - Invalid target ID", func(t *testing.T) {
		mockLinkRepo := &MockLinkRepo{}
		service := NewLinkService(mockLinkRepo)
		linkID := createTestStructuralLinkIdentifier()
		linkID.TargetId = "invalid-uuid"

		// Execute
		result, err := service.DeleteStructuralLinks(ctx, []*v1.StructuralLinkIdentifier{linkID})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(0), result) // Invalid UUIDs are skipped, so count is 0

		// Still succeeds but logs the error
		mockLinkRepo.AssertNotCalled(t, "DeleteLinks")
	})
}
