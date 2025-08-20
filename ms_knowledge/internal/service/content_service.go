package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"demo/ms_knowledge/internal/domain"
	"demo/ms_knowledge/internal/events"

	"github.com/google/uuid"
)

type ContentService struct {
	contentRepo  domain.ContentRepository
	spaceRepo    domain.SpaceRepository
	graphRepo    domain.GraphRepository
	storage      StorageService
	producer     *events.Producer
	sourceBucket string
	logger       *log.Logger
}

type StorageService interface {
	GeneratePresignedUploadURL(bucket, key string, expires time.Duration) (string, error)
	CalculateSHA256(data []byte) string
}

// NewContentService constructs the service. Pass nil producer if events are not needed (e.g., tests).
func NewContentService(contentRepo domain.ContentRepository, spaceRepo domain.SpaceRepository, graphRepo domain.GraphRepository, storage StorageService, producer *events.Producer, sourceBucket string, logger *log.Logger) *ContentService {
	return &ContentService{
		contentRepo:  contentRepo,
		spaceRepo:    spaceRepo,
		graphRepo:    graphRepo,
		storage:      storage,
		producer:     producer,
		sourceBucket: sourceBucket,
		logger:       logger,
	}
}

func (s *ContentService) CreateUploadURL(ctx context.Context, spaceID uuid.UUID, filename, mimeType string, sizeBytes int64, title string) (*domain.ContentSource, string, error) {
	// Validate inputs
	if spaceID == uuid.Nil {
		return nil, "", fmt.Errorf("space_id is required")
	}
	if strings.TrimSpace(filename) == "" {
		return nil, "", fmt.Errorf("filename is required")
	}
	if strings.TrimSpace(mimeType) == "" {
		return nil, "", fmt.Errorf("mime_type is required")
	}

	// Check if space exists
	exists, err := s.spaceRepo.Exists(ctx, spaceID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check space existence: %w", err)
	}
	if !exists {
		return nil, "", fmt.Errorf("space not found")
	}

	// Create content source record
	content := &domain.ContentSource{
		SpaceID:   spaceID,
		Status:    domain.ContentStatusUploading,
		MediaType: mimeType,
		Title:     strings.TrimSpace(title),
		Source:    filename,
		SizeBytes: sizeBytes,
	}

	err = s.contentRepo.Create(ctx, content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create content source: %w", err)
	}

	// Generate object key based on persisted content ID
	objectKey := fmt.Sprintf("spaces/%s/content/%s/%s", spaceID.String(), content.ID.String(), filename)

	// Generate pre-signed URL
	bucket := s.sourceBucket
	if bucket == "" {
		bucket = "knowledge-content"
	}
	uploadURL, err := s.storage.GeneratePresignedUploadURL(bucket, objectKey, 1*time.Hour)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return content, uploadURL, nil
}

func (s *ContentService) ConfirmUpload(ctx context.Context, contentID uuid.UUID, originalBlobHash string) (*domain.ContentSource, error) {
	// Get content source
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content source: %w", err)
	}

	// Update status to UPLOADED
	content.Status = domain.ContentStatusUploaded
	content.OriginalBlobHash = originalBlobHash

	err = s.contentRepo.Update(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to update content source: %w", err)
	}

	// Emit document.uploaded event
	if s.producer != nil {
		evt := events.DocumentUploadedEvent{
			ContentSourceID:  content.ID,
			OriginalBlobHash: originalBlobHash,
			SpaceID:          content.SpaceID,
		}
		if err := s.producer.ProduceJSON(ctx, events.TopicDocumentUploaded, content.ID.String(), evt); err != nil {
			s.logger.Printf("ERROR: failed to produce document.uploaded event for content_source_id (content.ID=%s): %v", content.ID, err)
			// Note: We don't return an error to the client here. The upload was confirmed
			// and the status is updated. The event failure should be handled by a
			// separate monitoring or reconciliation process.
		}
	}

	return content, nil
}

func (s *ContentService) GetContentSource(ctx context.Context, id uuid.UUID) (*domain.ContentSource, error) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get content source: %w", err)
	}

	return content, nil
}

func (s *ContentService) ListContentSources(ctx context.Context, filter domain.ContentSourceFilter) ([]*domain.ContentSource, error) {
	contents, err := s.contentRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list content sources: %w", err)
	}

	return contents, nil
}

func (s *ContentService) UpdateContentSourceStatus(ctx context.Context, id uuid.UUID, status domain.ContentStatus, processedBlobHash, errorMessage string) (*domain.ContentSource, error) {
	// Get content source
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get content source: %w", err)
	}

	// Update status
	err = s.contentRepo.UpdateStatus(ctx, id, status, processedBlobHash)
	if err != nil {
		return nil, fmt.Errorf("failed to update content source status: %w", err)
	}

	// Content nodes are automatically created in Neo4j when content is created in the repository
	// No additional graph operations needed here

	// Get updated content source
	content, err = s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated content source: %w", err)
	}

	return content, nil
}
