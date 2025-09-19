package python

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "demo/ms_canvas/go_app/api/proto/v1"
)

// ChunkEmbed performs combined chunking and embedding via the Python service
func (g *Gateway) ChunkEmbed(ctx context.Context, req *pb.ChunkEmbedRequest) (*pb.ChunkEmbedResponse, error) {
	// Call the Python service directly with protobuf types
	pbResp, err := g.client.ChunkEmbed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ChunkEmbed: %w", err)
	}

	return pbResp, nil
}

// EmbedQuery embeds a single query text
func (g *Gateway) EmbedQuery(ctx context.Context, text string, config *pb.EmbeddingConfig) (*pb.EmbedQueryResponse, error) {
	req := &pb.EmbedQueryRequest{
		Text:   text,
		Config: config,
	}

	return g.client.EmbedQuery(ctx, req)
}

// Healthz checks the health of the Python service
func (g *Gateway) Healthz(ctx context.Context) (*pb.HealthStatus, error) {
	return g.client.Healthz(ctx, &emptypb.Empty{})
}

// DefaultChunkingConfig returns default chunking configuration
func DefaultChunkingConfig() *pb.ChunkingConfig {
	return &pb.ChunkingConfig{
		Type:           "sentence",
		TargetTokens:   300,
		OverlapPercent: 10,
		Tokenizer:      "tiktoken:cl100k_base",
	}
}

// DefaultEmbeddingConfig returns default embedding configuration
func DefaultEmbeddingConfig() *pb.EmbeddingConfig {
	return &pb.EmbeddingConfig{
		Provider:     "huggingface",
		ModelId:      "all-MiniLM-L6-v2",
		ModelVersion: "",
	}
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
