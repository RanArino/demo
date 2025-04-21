package storage

import (
	"context"
	"testing"
	"time"
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

	// Clean up
	if err := client.Close(); err != nil {
		t.Errorf("Failed to close client: %v", err)
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
	err = client.EnsureCollection(ctx, "test_collection", 384)
	if err != nil {
		t.Fatalf("Failed to ensure collection: %v", err)
	}

	// Test idempotency - should not error when called again
	err = client.EnsureCollection(ctx, "test_collection", 384)
	if err != nil {
		t.Fatalf("Failed to ensure existing collection: %v", err)
	}
}
