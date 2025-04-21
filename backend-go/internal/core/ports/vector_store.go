package ports

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
)

// VectorStore defines operations for vector storage and search
type VectorStore interface {
	// EnsureCollection ensures a collection exists with the specified parameters
	EnsureCollection(ctx context.Context, collectionName string, vectorSize uint64, distance qdrant.Distance) error

	// UpsertPoints adds or updates points in a collection
	UpsertPoints(ctx context.Context, collectionName string, points []*qdrant.PointStruct) error

	// Search performs a similarity search in the collection
	Search(ctx context.Context, collectionName string, vector []float32, limit uint64, filter *qdrant.Filter) ([]*qdrant.ScoredPoint, error)

	// DeletePoints removes points from a collection
	DeletePoints(ctx context.Context, collectionName string, pointIDs []string) error

	// GetPoints retrieves points by their IDs
	GetPoints(ctx context.Context, collectionName string, pointIDs []string) ([]*qdrant.RetrievedPoint, error)

	// CountPoints returns the total number of points in a collection
	CountPoints(ctx context.Context, collectionName string) (uint64, error)
}

// Note: This interface is implemented by internal/storage/qdrant.go QdrantClient
// The implementation is validated by tests in internal/storage/qdrant_test.go
