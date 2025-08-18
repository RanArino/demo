package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ContentStatus represents the processing status of a content source
type ContentStatus string

const (
	ContentStatusUploading  ContentStatus = "UPLOADING"
	ContentStatusUploaded   ContentStatus = "UPLOADED"
	ContentStatusProcessing ContentStatus = "PROCESSING"
	ContentStatusProcessed  ContentStatus = "PROCESSED"
	ContentStatusFailed     ContentStatus = "FAILED"
)

// ContentSource represents a content source in the system
type ContentSource struct {
	ID                uuid.UUID     `json:"id"`
	SpaceID           uuid.UUID     `json:"space_id"`
	OwnerID           uuid.UUID     `json:"owner_id"`
	Title             string        `json:"title"`
	MediaType         string        `json:"media_type"`
	Source            string        `json:"source"`
	SizeBytes         int64         `json:"size_bytes"`
	Status            ContentStatus `json:"status"`
	OriginalBlobHash  string        `json:"original_blob_hash"`
	ProcessedBlobHash *string       `json:"processed_blob_hash"`
	ContentSummary    *string       `json:"content_summary"`
	Keywords          []string      `json:"keywords"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	DeletedAt         *time.Time    `json:"deleted_at"`
}

// ContentSourceFilter represents filters for listing content sources
type ContentSourceFilter struct {
	SpaceID       uuid.UUID
	Title         string
	Query         string // Searches title and content_summary
	MediaType     string
	Keywords      []string
	Status        ContentStatus
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
	Limit         int
	Offset        int
}

// ContentRepository defines the interface for content source data operations
type ContentRepository interface {
	Create(ctx context.Context, content *ContentSource) error
	GetByID(ctx context.Context, id uuid.UUID) (*ContentSource, error)
	List(ctx context.Context, filter ContentSourceFilter) ([]*ContentSource, error)
	Update(ctx context.Context, content *ContentSource) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status ContentStatus, processedBlobHash string) error
	Delete(ctx context.Context, id uuid.UUID) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	CountBySpace(ctx context.Context, spaceID uuid.UUID) (int64, error)
}

// KnowledgeContentBlob represents a content-addressable storage registry entry
type KnowledgeContentBlob struct {
	BlobHash      string    `json:"blob_hash"`
	StorageBucket string    `json:"storage_bucket"`
	StorageKey    string    `json:"storage_key"`
	SizeBytes     int64     `json:"size_bytes"`
	CreatedAt     time.Time `json:"created_at"`
}

// BlobRepository defines the interface for blob registry operations
type BlobRepository interface {
	Create(ctx context.Context, blob *KnowledgeContentBlob) error
	GetByHash(ctx context.Context, blobHash string) (*KnowledgeContentBlob, error)
	Exists(ctx context.Context, blobHash string) (bool, error)
	Delete(ctx context.Context, blobHash string) error
}
