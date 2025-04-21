package ports

import (
	"context"

	"github.com/ran/demo/backend-go/internal/core/models"
)

// DocumentRepository defines operations for Document persistence
type DocumentRepository interface {
	// Save persists a Document
	Save(ctx context.Context, doc *models.Document) error

	// GetByID retrieves a Document by its ID
	GetByID(ctx context.Context, id string) (*models.Document, error)

	// UpdateStatus updates the processing status of a Document
	UpdateStatus(ctx context.Context, id string, status models.ProcessingStatus, err error) error

	// List retrieves all Documents with optional filtering
	List(ctx context.Context) ([]*models.Document, error)
}

// ChunkRepository defines operations for Chunk persistence
type ChunkRepository interface {
	// SaveBatch persists multiple Chunks
	SaveBatch(ctx context.Context, chunks []*models.Chunk) error

	// GetByDocumentID retrieves all Chunks for a Document
	GetByDocumentID(ctx context.Context, documentID string) ([]*models.Chunk, error)

	// GetByID retrieves a Chunk by its ID
	GetByID(ctx context.Context, id string) (*models.Chunk, error)
}

// SummaryRepository defines operations for Summary persistence
type SummaryRepository interface {
	// Save persists a Summary
	Save(ctx context.Context, summary *models.Summary) error

	// GetByDocumentID retrieves the Summary for a Document
	GetByDocumentID(ctx context.Context, documentID string) (*models.Summary, error)

	// GetByID retrieves a Summary by its ID
	GetByID(ctx context.Context, id string) (*models.Summary, error)
}
