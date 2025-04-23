package ports

import (
	"context"

	"github.com/ran/demo/backend-go/internal/core/models"
)

// Use core models for domain entities

// DocumentUploader handles file ingestion
type DocumentUploader interface {
	Upload(ctx context.Context, filePaths []string) ([]models.Document, error)
}

// VectorStoreService defines operations for segmenting text, indexing vectors, and searching.
type VectorStoreService interface {
	// Segment splits a Document into smaller Chunks.
	Segment(ctx context.Context, doc models.Document, maxTokens int) ([]models.Chunk, error)
	// Index indexes embeddings and metadata into a storage backend.
	Index(ctx context.Context, collection string, id string, vector []float32, meta map[string]interface{}) error
	// Search performs similarity search over indexed vectors.
	Search(ctx context.Context, collection string, vector []float32, topK int) ([]models.SearchResult, error)
}

// VectorAnalysisService provides vector analysis capabilities such as dimensionality reduction and clustering for visualization and grouping.
type VectorAnalysisService interface {
	// Reduce reduces high-dimensional vectors for visualization (e.g., UMAP, PCA).
	Reduce(ctx context.Context, vectors [][]float32) ([][]float32, error)
	// Cluster groups vectors into clusters for visual distinction (e.g., K-Means, HDBSCAN).
	Cluster(ctx context.Context, vectors [][]float32) ([]models.Cluster, error)
}
