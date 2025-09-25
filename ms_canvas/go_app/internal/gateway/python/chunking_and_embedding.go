package python

import (
	"context"
	"fmt"
	"strconv"

	privpb "demo/ms_canvas/go_app/api/proto/private/v1"
	"demo/ms_canvas/go_app/internal/config"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ChunkEmbed performs combined chunking and embedding via the Python service
func (g *Gateway) ChunkEmbed(ctx context.Context, req *privpb.ChunkEmbedRequest) (*privpb.ChunkEmbedResponse, error) {
	// Call the Python service directly with protobuf types
	pbResp, err := g.client.ChunkEmbed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ChunkEmbed: %w", err)
	}

	return pbResp, nil
}

// ChunkDocumentRequest represents a request to chunk a document
type ChunkDocumentRequest struct {
	DocumentContent string
	ContentNodeID   string
	SpaceID         string
}

// ChunkDocumentResponse represents the response from chunking a document
type ChunkDocumentResponse struct {
	Chunks []ChunkInfo
}

// ChunkInfo represents information about a single chunk
type ChunkInfo struct {
	ID            string
	Content       string
	Embedding     []float32
	SequenceIndex int32
	StartPosition int64
	EndPosition   int64
}

// ChunkDocument processes a document through chunking and returns chunk information
func (g *Gateway) ChunkDocument(ctx context.Context, documentContent []byte, contentNodeID string) (*ChunkDocumentResponse, error) {
	// Convert document content to string
	contentStr := string(documentContent)

	// Create chunk embed request using text source
	req := &privpb.ChunkEmbedRequest{
		SpaceId:       contentNodeID, // Use content node ID as space ID for now
		ContentNodeId: contentNodeID,
		Source:        &privpb.ChunkEmbedRequest_Text{Text: contentStr},
		Chunking:      DefaultChunkingConfig(),
		Embedding:     DefaultGeminiEmbeddingConfig(),
		BatchSize:     0, // Default batch size
	}

	// Call the Python service
	resp, err := g.ChunkEmbed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk document: %w", err)
	}

	// Convert response to our format
	chunks := make([]ChunkInfo, len(resp.Results))
	for i, chunkEmbedding := range resp.Results {
		chunks[i] = ChunkInfo{
			ID:            chunkEmbedding.Chunk.Id,
			Content:       chunkEmbedding.Chunk.Content,
			Embedding:     chunkEmbedding.Vector,
			SequenceIndex: chunkEmbedding.Chunk.SequenceIndex,
			StartPosition: chunkEmbedding.Chunk.StartPosition,
			EndPosition:   chunkEmbedding.Chunk.EndPosition,
		}
	}

	return &ChunkDocumentResponse{
		Chunks: chunks,
	}, nil
}

// EmbedQuery embeds a single query text
func (g *Gateway) EmbedQuery(ctx context.Context, text string, config *privpb.EmbeddingConfig) (*privpb.EmbedQueryResponse, error) {
	req := &privpb.EmbedQueryRequest{
		Text:   text,
		Config: config,
	}

	return g.client.EmbedQuery(ctx, req)
}

// Healthz checks the health of the Python service
func (g *Gateway) Healthz(ctx context.Context) (*privpb.HealthStatus, error) {
	return g.client.Healthz(ctx, &emptypb.Empty{})
}

// DefaultChunkingConfig returns default chunking configuration
func DefaultChunkingConfig() *privpb.ChunkingConfig {
	return &privpb.ChunkingConfig{
		Type:           "sentence",
		TargetTokens:   300,
		OverlapPercent: 10,
		Tokenizer:      "tiktoken:cl100k_base",
	}
}

// DefaultEmbeddingConfig returns default embedding configuration
func DefaultEmbeddingConfig() *privpb.EmbeddingConfig {
	return &privpb.EmbeddingConfig{
		Provider:     "huggingface",
		ModelId:      "all-MiniLM-L6-v2",
		ModelVersion: "",
	}
}

// DefaultGeminiEmbeddingConfig returns default Gemini embedding configuration
func DefaultGeminiEmbeddingConfig() *privpb.EmbeddingConfig {
	return &privpb.EmbeddingConfig{
		Provider:             "gemini",
		ModelId:              "gemini-embedding-001",
		ModelVersion:         "",
		OutputDimensionality: 1536,
		TaskType:             "SEMANTIC_SIMILARITY",
		NormalizeEmbeddings:  true,
	}
}

// GetGeminiAPIKey returns the Gemini API key from configuration
func GetGeminiAPIKey(cfg *config.Config) string {
	return cfg.GEMINIAPIKey
}

// ParseBatchSize parses batch size from string, returns nil if empty or invalid
func ParseBatchSize(s string) *int32 {
	if s == "" {
		return nil
	}
	if val, err := strconv.ParseInt(s, 10, 32); err == nil && val > 0 {
		result := int32(val)
		return &result
	}
	return nil
}
