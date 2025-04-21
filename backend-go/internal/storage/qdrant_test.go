package storage

import (
	"context"
	"testing"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"github.com/ran/demo/backend-go/internal/core/models"
)

func TestNewQdrantClient(t *testing.T) {
	// This test doesn't actually connect to Qdrant
	client, err := NewQdrantClient("localhost:6334")
	if err != nil {
		t.Fatalf("Failed to create Qdrant client: %v", err)
	}

	// Verify we have both clients initialized
	if client.client == nil {
		t.Error("QdrantClient's client field is nil")
	}

	if client.collections == nil {
		t.Error("QdrantClient's collections field is nil")
	}

	if client.points == nil {
		t.Error("QdrantClient's points field is nil")
	}

	// Clean up
	if err := client.Close(); err != nil {
		t.Errorf("Failed to close client: %v", err)
	}
}

func TestCreatePointFromChunk(t *testing.T) {
	chunk := &models.Chunk{
		ID:         "test-chunk-id",
		DocumentID: "test-doc-id",
		Index:      1,
		Text:       "This is a test chunk",
		TokenCount: 5,
		Embedding:  []float32{0.1, 0.2, 0.3},
	}

	point, err := CreatePointFromChunk(chunk)
	if err != nil {
		t.Fatalf("Failed to create point from chunk: %v", err)
	}

	// Check ID
	if point.Id.GetUuid() != chunk.ID {
		t.Errorf("Expected point ID %s, got %s", chunk.ID, point.Id.GetUuid())
	}

	// Check vector
	if len(point.Vectors.GetVector().Data) != len(chunk.Embedding) {
		t.Errorf("Expected vector length %d, got %d", len(chunk.Embedding), len(point.Vectors.GetVector().Data))
	}

	// Check payload
	if point.Payload["document_id"].GetStringValue() != chunk.DocumentID {
		t.Errorf("Expected document_id %s, got %s", chunk.DocumentID, point.Payload["document_id"].GetStringValue())
	}

	if point.Payload["text"].GetStringValue() != chunk.Text {
		t.Errorf("Expected text %s, got %s", chunk.Text, point.Payload["text"].GetStringValue())
	}

	if point.Payload["index"].GetIntegerValue() != int64(chunk.Index) {
		t.Errorf("Expected index %d, got %d", chunk.Index, point.Payload["index"].GetIntegerValue())
	}
}

func TestCreatePointFromSummary(t *testing.T) {
	summary := &models.Summary{
		ID:         "test-summary-id",
		DocumentID: "test-doc-id",
		Text:       "This is a test summary",
		Embedding:  []float32{0.4, 0.5, 0.6},
	}

	point, err := CreatePointFromSummary(summary)
	if err != nil {
		t.Fatalf("Failed to create point from summary: %v", err)
	}

	// Check ID
	if point.Id.GetUuid() != summary.ID {
		t.Errorf("Expected point ID %s, got %s", summary.ID, point.Id.GetUuid())
	}

	// Check vector
	if len(point.Vectors.GetVector().Data) != len(summary.Embedding) {
		t.Errorf("Expected vector length %d, got %d", len(summary.Embedding), len(point.Vectors.GetVector().Data))
	}

	// Check payload
	if point.Payload["document_id"].GetStringValue() != summary.DocumentID {
		t.Errorf("Expected document_id %s, got %s", summary.DocumentID, point.Payload["document_id"].GetStringValue())
	}

	if point.Payload["text"].GetStringValue() != summary.Text {
		t.Errorf("Expected text %s, got %s", summary.Text, point.Payload["text"].GetStringValue())
	}

	if !point.Payload["is_summary"].GetBoolValue() {
		t.Errorf("Expected is_summary to be true, got false")
	}
}

// Skip integration test unless QDRANT_INTEGRATION_TEST env var is set
func TestEnsureCollection_Integration(t *testing.T) {
	t.Skip("Skipping integration test. Set QDRANT_INTEGRATION_TEST=1 to run.")

	client, err := NewQdrantClient("localhost:6334")
	if err != nil {
		t.Fatalf("Failed to create Qdrant client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test collection creation
	err = client.EnsureCollection(ctx, "test_collection", 384, qdrant.Distance_Cosine)
	if err != nil {
		t.Fatalf("Failed to ensure collection: %v", err)
	}

	// Test idempotency - should not error when called again
	err = client.EnsureCollection(ctx, "test_collection", 384, qdrant.Distance_Cosine)
	if err != nil {
		t.Fatalf("Failed to ensure existing collection: %v", err)
	}
}
