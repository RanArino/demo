package qdrant

import (
	"context"
	"testing"

	sdk "github.com/qdrant/go-client/qdrant"
	"github.com/ran/demo/backend-go/internal/config"
)

// getQdrantEnv loads QDRANT_HOST and QDRANT_API_KEY
func getQdrantEnv(t *testing.T) (string, string) {
	host := config.GetEnv("QDRANT_HOST")
	apiKey := config.GetEnv("QDRANT_API_KEY")
	if host == "" || apiKey == "" {
		t.Skip("QDRANT_HOST and QDRANT_API_KEY must be set for integration test")
	}
	return host, apiKey
}

func TestQdrantClientConnection(t *testing.T) {
	host, apiKey := getQdrantEnv(t)
	client, err := NewQdrantClient(host, apiKey)
	if err != nil {
		t.Fatalf("failed to create Qdrant client: %v", err)
	}
	defer client.Close()
}

func TestQdrantEnsureCollection(t *testing.T) {
	host, apiKey := getQdrantEnv(t)
	client, err := NewQdrantClient(host, apiKey)
	if err != nil {
		t.Fatalf("failed to create Qdrant client: %v", err)
	}
	defer client.Close()

	// // Existing collection
	collectionName := "midjourney"
	// New collection
	// collectionName := "cascade_test_collection"
	err = client.EnsureCollection(context.Background(), collectionName, 512, sdk.Distance_Cosine)
	if err != nil {
		t.Fatalf("EnsureCollection failed: %v", err)
	}
}
