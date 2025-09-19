package python

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "demo/ms_canvas/go_app/api/proto/v1"
)

// MockCanvasInternalClient is a mock implementation of pb.CanvasInternalClient
type MockCanvasInternalClient struct {
	mock.Mock
}

func (m *MockCanvasInternalClient) ChunkEmbed(ctx context.Context, req *pb.ChunkEmbedRequest, opts ...grpc.CallOption) (*pb.ChunkEmbedResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.ChunkEmbedResponse), args.Error(1)
}

func (m *MockCanvasInternalClient) EmbedQuery(ctx context.Context, req *pb.EmbedQueryRequest, opts ...grpc.CallOption) (*pb.EmbedQueryResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.EmbedQueryResponse), args.Error(1)
}

func (m *MockCanvasInternalClient) Healthz(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*pb.HealthStatus, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.HealthStatus), args.Error(1)
}

// Helper function to create a gateway with mock client
func newTestGateway() (*Gateway, *MockCanvasInternalClient) {
	mockClient := &MockCanvasInternalClient{}
	gateway := &Gateway{
		client: mockClient,
		conn:   nil, // Not needed for unit tests
	}
	return gateway, mockClient
}

func TestGateway_ChunkEmbed_Success(t *testing.T) {
	gateway, mockClient := newTestGateway()
	ctx := context.Background()

	// Prepare test data
	req := &pb.ChunkEmbedRequest{
		SpaceId:         "space-123",
		ContentSourceId: "content-456",
		ContentNodeId:   "node-789",
		Source: &pb.ChunkEmbedRequest_Text{
			Text: "This is test content for chunking and embedding.",
		},
		Chunking: &pb.ChunkingConfig{
			Type:           "sentence",
			TargetTokens:   300,
			OverlapPercent: 10,
			Tokenizer:      "tiktoken:cl100k_base",
		},
		Embedding: &pb.EmbeddingConfig{
			Provider:     "huggingface",
			ModelId:      "all-MiniLM-L6-v2",
			ModelVersion: "",
		},
		BatchSize: 50,
	}

	expectedResponse := &pb.ChunkEmbedResponse{
		Results: []*pb.ChunkEmbedding{
			{
				Chunk: &pb.Chunk{
					Id:            "chunk-1",
					SequenceIndex: 0,
					StartPosition: 0,
					EndPosition:   48,
					Content:       "This is test content for chunking and embedding.",
				},
				Vector: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
			},
		},
		Dims:         5,
		ModelId:      "all-MiniLM-L6-v2",
		ModelVersion: "v1.0",
	}

	// Set up mock expectations
	mockClient.On("ChunkEmbed", ctx, req).Return(expectedResponse, nil)

	// Execute
	result, err := gateway.ChunkEmbed(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestGateway_ChunkEmbed_Error(t *testing.T) {
	gateway, mockClient := newTestGateway()
	ctx := context.Background()

	req := &pb.ChunkEmbedRequest{
		SpaceId:         "space-123",
		ContentSourceId: "content-456",
		Source: &pb.ChunkEmbedRequest_Text{
			Text: "Test content",
		},
	}

	expectedError := errors.New("python service unavailable")

	// Set up mock expectations
	mockClient.On("ChunkEmbed", ctx, req).Return((*pb.ChunkEmbedResponse)(nil), expectedError)

	// Execute
	result, err := gateway.ChunkEmbed(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to call ChunkEmbed")
	assert.Contains(t, err.Error(), "python service unavailable")
	mockClient.AssertExpectations(t)
}

func TestGateway_EmbedQuery_Success(t *testing.T) {
	gateway, mockClient := newTestGateway()
	ctx := context.Background()

	text := "What is the meaning of life?"
	config := &pb.EmbeddingConfig{
		Provider:     "openai",
		ModelId:      "text-embedding-ada-002",
		ModelVersion: "v2",
	}

	expectedRequest := &pb.EmbedQueryRequest{
		Text:   text,
		Config: config,
	}

	expectedResponse := &pb.EmbedQueryResponse{
		Vector:       []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		Dims:         5,
		ModelId:      "text-embedding-ada-002",
		ModelVersion: "v2",
	}

	// Set up mock expectations
	mockClient.On("EmbedQuery", ctx, expectedRequest).Return(expectedResponse, nil)

	// Execute
	result, err := gateway.EmbedQuery(ctx, text, config)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestGateway_EmbedQuery_Error(t *testing.T) {
	gateway, mockClient := newTestGateway()
	ctx := context.Background()

	text := "Test query"
	config := &pb.EmbeddingConfig{
		Provider: "openai",
		ModelId:  "text-embedding-ada-002",
	}

	expectedRequest := &pb.EmbedQueryRequest{
		Text:   text,
		Config: config,
	}

	expectedError := errors.New("embedding service error")

	// Set up mock expectations
	mockClient.On("EmbedQuery", ctx, expectedRequest).Return((*pb.EmbedQueryResponse)(nil), expectedError)

	// Execute
	result, err := gateway.EmbedQuery(ctx, text, config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockClient.AssertExpectations(t)
}

func TestGateway_Healthz_Success(t *testing.T) {
	gateway, mockClient := newTestGateway()
	ctx := context.Background()

	expectedResponse := &pb.HealthStatus{
		Status: "OK",
		Components: map[string]string{
			"chunking":  "healthy",
			"embedding": "healthy",
		},
	}

	// Set up mock expectations
	mockClient.On("Healthz", ctx, &emptypb.Empty{}).Return(expectedResponse, nil)

	// Execute
	result, err := gateway.Healthz(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestGateway_Healthz_Error(t *testing.T) {
	gateway, mockClient := newTestGateway()
	ctx := context.Background()

	expectedError := errors.New("health check failed")

	// Set up mock expectations
	mockClient.On("Healthz", ctx, &emptypb.Empty{}).Return((*pb.HealthStatus)(nil), expectedError)

	// Execute
	result, err := gateway.Healthz(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockClient.AssertExpectations(t)
}

func TestDefaultChunkingConfig(t *testing.T) {
	config := DefaultChunkingConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "sentence", config.Type)
	assert.Equal(t, int32(300), config.TargetTokens)
	assert.Equal(t, int32(10), config.OverlapPercent)
	assert.Equal(t, "tiktoken:cl100k_base", config.Tokenizer)
}

func TestDefaultEmbeddingConfig(t *testing.T) {
	config := DefaultEmbeddingConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "huggingface", config.Provider)
	assert.Equal(t, "all-MiniLM-L6-v2", config.ModelId)
	assert.Equal(t, "", config.ModelVersion)
}

func TestParseBatchSize(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *int32
	}{
		{
			name:     "valid positive number",
			input:    "50",
			expected: func() *int32 { v := int32(50); return &v }(),
		},
		{
			name:     "valid large number",
			input:    "1000",
			expected: func() *int32 { v := int32(1000); return &v }(),
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "zero",
			input:    "0",
			expected: nil,
		},
		{
			name:     "negative number",
			input:    "-10",
			expected: nil,
		},
		{
			name:     "invalid string",
			input:    "abc",
			expected: nil,
		},
		{
			name:     "float number",
			input:    "50.5",
			expected: nil,
		},
		{
			name:     "number with spaces",
			input:    " 50 ",
			expected: nil,
		},
		{
			name:     "very large number beyond int32",
			input:    "999999999999",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseBatchSize(tc.input)

			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tc.expected, *result)
			}
		})
	}
}

// TestChunkEmbedWith各种参数组合
func TestGateway_ChunkEmbed_WithDifferentSources(t *testing.T) {
	gateway, mockClient := newTestGateway()
	ctx := context.Background()

	testCases := []struct {
		name string
		req  *pb.ChunkEmbedRequest
	}{
		{
			name: "with text source",
			req: &pb.ChunkEmbedRequest{
				SpaceId:         "space-123",
				ContentSourceId: "content-456",
				Source: &pb.ChunkEmbedRequest_Text{
					Text: "Test content with text source",
				},
				Chunking:  DefaultChunkingConfig(),
				Embedding: DefaultEmbeddingConfig(),
			},
		},
		{
			name: "with blob URL source",
			req: &pb.ChunkEmbedRequest{
				SpaceId:         "space-123",
				ContentSourceId: "content-456",
				Source: &pb.ChunkEmbedRequest_BlobUrl{
					BlobUrl: "https://example.com/document.txt",
				},
				Chunking:  DefaultChunkingConfig(),
				Embedding: DefaultEmbeddingConfig(),
			},
		},
		{
			name: "with batch size",
			req: &pb.ChunkEmbedRequest{
				SpaceId:         "space-123",
				ContentSourceId: "content-456",
				Source: &pb.ChunkEmbedRequest_Text{
					Text: "Test content with batch size",
				},
				Chunking:  DefaultChunkingConfig(),
				Embedding: DefaultEmbeddingConfig(),
				BatchSize: 25,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedResponse := &pb.ChunkEmbedResponse{
				Results:      []*pb.ChunkEmbedding{},
				Dims:         1536,
				ModelId:      "all-MiniLM-L6-v2",
				ModelVersion: "v1.0",
			}

			mockClient.On("ChunkEmbed", ctx, tc.req).Return(expectedResponse, nil).Once()

			result, err := gateway.ChunkEmbed(ctx, tc.req)

			assert.NoError(t, err)
			assert.Equal(t, expectedResponse, result)
		})
	}

	mockClient.AssertExpectations(t)
}
