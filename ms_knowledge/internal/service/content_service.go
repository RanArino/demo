package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

type ContentService struct {
	contentRepo domain.ContentRepository
	spaceRepo   domain.SpaceRepository
	graphRepo   domain.GraphRepository
	storage     StorageService
}

type StorageService interface {
	GeneratePresignedUploadURL(bucket, key string, expires time.Duration) (string, error)
	CalculateSHA256(data []byte) string
}

func NewContentService(contentRepo domain.ContentRepository, spaceRepo domain.SpaceRepository, graphRepo domain.GraphRepository, storage StorageService) *ContentService {
	return &ContentService{
		contentRepo: contentRepo,
		spaceRepo:   spaceRepo,
		graphRepo:   graphRepo,
		storage:     storage,
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

	// Generate object key
	objectKey := fmt.Sprintf("spaces/%s/content/%s/%s", spaceID.String(), uuid.New().String(), filename)

	// Create content source record
	content := &domain.ContentSource{
		SpaceID:   spaceID,
		Status:    domain.ContentStatusUploading,
		MediaType: mimeType,
		Title:     strings.TrimSpace(title),
		Source:    filename,
	}

	err = s.contentRepo.Create(ctx, content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create content source: %w", err)
	}

	// Generate pre-signed URL
	uploadURL, err := s.storage.GeneratePresignedUploadURL("knowledge-content", objectKey, 1*time.Hour)
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

	// TODO: Emit document.uploaded event to Kafka
	// This will be implemented when Kafka integration is added

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
