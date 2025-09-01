package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"demo/ms_knowledge/internal/config"
	"demo/ms_knowledge/internal/domain"
	"demo/ms_knowledge/internal/events"

	"github.com/google/uuid"
)

type EventProducer interface {
	ProduceJSON(ctx context.Context, topic string, key string, v any) error
	Close()
}

type ContentService struct {
	contentRepo domain.ContentRepository
	spaceRepo   domain.SpaceRepository
	graphRepo   domain.GraphRepository
	storage     StorageService
	producer    EventProducer
	cfg         *config.Config
	logger      *log.Logger
}

type StorageService interface {
	GeneratePresignedUploadURL(bucket, key string, expires time.Duration) (string, error)
	CalculateSHA256(data []byte) string
	GeneratePresignedDownloadURL(bucket, key string, expires time.Duration) (string, error)
	DeleteObject(bucket, key string) error
}

// NewContentService constructs the service. Pass nil producer if events are not needed (e.g., tests).
func NewContentService(contentRepo domain.ContentRepository, spaceRepo domain.SpaceRepository, graphRepo domain.GraphRepository, storage StorageService, producer EventProducer, cfg *config.Config, logger *log.Logger) *ContentService {
	return &ContentService{
		contentRepo: contentRepo,
		spaceRepo:   spaceRepo,
		graphRepo:   graphRepo,
		storage:     storage,
		producer:    producer,
		cfg:         cfg,
		logger:      logger,
	}
}

// scopeContentFilterToOwner forces the filter to the caller's owner_id.
func (s *ContentService) scopeContentFilterToOwner(ctx context.Context, filter domain.ContentSourceFilter) (domain.ContentSourceFilter, error) {
	ownerUUID, err := MustGetOwnerUUID(ctx)
	if err != nil {
		return filter, err
	}
	filter.OwnerID = ownerUUID
	return filter, nil
}

// resolveBucket returns the bucket name for a given kind ("source" or "processed").
func (s *ContentService) resolveBucket(kind string) (string, error) {
	// Prefer unified content-source bucket if configured
	if s.cfg != nil && strings.TrimSpace(s.cfg.R2.BucketContentSourceName) != "" {
		return s.cfg.R2.BucketContentSourceName, nil
	}
	// Legacy fallback by kind
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "source", "original":
		if s.cfg == nil || s.cfg.R2.BucketSourceName == "" {
			return "", fmt.Errorf("R2 source bucket not configured")
		}
		return s.cfg.R2.BucketSourceName, nil
	case "processed":
		if s.cfg == nil || s.cfg.R2.BucketProcessedName == "" {
			return "", fmt.Errorf("R2 processed bucket not configured")
		}
		return s.cfg.R2.BucketProcessedName, nil
	default:
		return "", fmt.Errorf("invalid kind: %s", kind)
	}
}

// processedFilename derives the processed filename by replacing the extension with .md or appending .md
func processedFilename(original string) string {
	name := strings.TrimSpace(original)
	if name == "" {
		return "processed.md"
	}
	if dot := strings.LastIndex(name, "."); dot > 0 {
		return name[:dot] + ".md"
	}
	return name + ".md"
}

// buildObjectKey constructs the R2 object key including the owner prefix and space/content path
func buildObjectKey(ownerID, spaceID, contentID uuid.UUID, filename string) string {
	// <userId>/spaces/<spaceId>/content/<contentSourceId>/<fileName>
	return ownerID.String() + "/spaces/" + spaceID.String() + "/content/" + contentID.String() + "/" + strings.TrimSpace(filename)
}

func (s *ContentService) CreateUploadURL(ctx context.Context, spaceID uuid.UUID, filename, mimeType string, sizeBytes int64, title string) (*domain.ContentSource, string, error) {
	content, uploadURL, _, _, err := s.CreateUploadURLWithKind(ctx, spaceID, filename, mimeType, sizeBytes, title, "source")
	if err != nil {
		return nil, "", err
	}
	return content, uploadURL, nil
}

// CreateUploadURLWithKind issues a presigned upload URL for the requested kind ("source" or "processed").
// Returns content, uploadURL, objectKey, expiresAt.
func (s *ContentService) CreateUploadURLWithKind(ctx context.Context, spaceID uuid.UUID, filename, mimeType string, sizeBytes int64, title, kind string) (*domain.ContentSource, string, string, time.Time, error) {
	// Validate inputs
	if spaceID == uuid.Nil {
		return nil, "", "", time.Time{}, fmt.Errorf("space_id is required")
	}
	if strings.TrimSpace(filename) == "" {
		return nil, "", "", time.Time{}, fmt.Errorf("filename is required")
	}
	if strings.TrimSpace(mimeType) == "" {
		return nil, "", "", time.Time{}, fmt.Errorf("mime_type is required")
	}
	if strings.TrimSpace(kind) == "" {
		return nil, "", "", time.Time{}, fmt.Errorf("kind is required")
	}

	// Check if space exists
	exists, err := s.spaceRepo.Exists(ctx, spaceID)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("failed to check space existence: %w", err)
	}
	if !exists {
		return nil, "", "", time.Time{}, fmt.Errorf("space not found")
	}

	// Derive owner for the content from context
	ownerUUID, err := MustGetOwnerUUID(ctx)
	if err != nil {
		return nil, "", "", time.Time{}, err
	}

	// Derive title: if request title is empty, default to filename
	effectiveTitle := strings.TrimSpace(title)
	if effectiveTitle == "" {
		effectiveTitle = filename
	}

	// Create content source record
	content := &domain.ContentSource{
		SpaceID:   spaceID,
		OwnerID:   ownerUUID,
		Status:    domain.ContentStatusUploading,
		MediaType: mimeType,
		Title:     effectiveTitle,
		Source:    filename,
		SizeBytes: sizeBytes,
	}

	err = s.contentRepo.Create(ctx, content)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("failed to create content source: %w", err)
	}

	// Generate object key based on persisted content ID
	objectKey := buildObjectKey(ownerUUID, spaceID, content.ID, filename)

	// Resolve bucket for requested kind
	bucket, err := s.resolveBucket(kind)
	if err != nil {
		return nil, "", "", time.Time{}, err
	}

	// Short TTL (default 15m)
	expires := 15 * time.Minute
	uploadURL, err := s.storage.GeneratePresignedUploadURL(bucket, objectKey, expires)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return content, uploadURL, objectKey, time.Now().Add(expires), nil
}

// ConfirmUpload remains for backward-compatibility and updates original blob hash.
func (s *ContentService) ConfirmUpload(ctx context.Context, contentID uuid.UUID, originalBlobHash string) (*domain.ContentSource, error) {
	return s.ConfirmUploadWithKind(ctx, contentID, "source", originalBlobHash)
}

// ConfirmUploadWithKind updates only the corresponding hash field based on kind.
func (s *ContentService) ConfirmUploadWithKind(ctx context.Context, contentID uuid.UUID, kind string, blobHash string) (*domain.ContentSource, error) {
	// Get content source
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content source: %w", err)
	}

	// Enforce RLS: caller must be the owner
	if err := EnforceOwner(ctx, content.OwnerID); err != nil {
		return nil, err
	}

	// Update only the targeted field
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "source", "original":
		content.Status = domain.ContentStatusUploaded
		content.OriginalBlobHash = blobHash
	case "processed":
		content.ProcessedBlobHash = &blobHash
	default:
		return nil, fmt.Errorf("invalid kind: %s", kind)
	}

	err = s.contentRepo.Update(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to update content source: %w", err)
	}

	// Emit document.uploaded event only for ORIGINAL confirms
	if s.producer != nil && (strings.ToLower(strings.TrimSpace(kind)) == "source" || strings.ToLower(strings.TrimSpace(kind)) == "original") {
		objectKey := buildObjectKey(content.OwnerID, content.SpaceID, content.ID, strings.TrimSpace(content.Source))
		evt := events.DocumentUploadedEvent{
			ContentSourceID:   content.ID,
			OriginalBlobHash:  blobHash,
			SpaceID:           content.SpaceID,
			OriginalObjectKey: objectKey,
		}
		if err := s.producer.ProduceJSON(ctx, events.TopicDocumentUploaded, content.ID.String(), evt); err != nil {
			s.logger.Printf("ERROR: failed to produce document.uploaded event for content_source_id (content.ID=%s): %v", content.ID, err)
		}
	}

	return content, nil
}

func (s *ContentService) GetContentSource(ctx context.Context, id uuid.UUID) (*domain.ContentSource, error) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get content source: %w", err)
	}

	// Enforce RLS
	if err := EnforceOwner(ctx, content.OwnerID); err != nil {
		return nil, err
	}

	return content, nil
}

func (s *ContentService) ListContentSources(ctx context.Context, filter domain.ContentSourceFilter) ([]*domain.ContentSource, error) {
	// Enforce RLS: always scope to caller's owner_id
	var err error
	filter, err = s.scopeContentFilterToOwner(ctx, filter)
	if err != nil {
		return nil, err
	}

	// If filtering by specific space, validate it exists
	if filter.SpaceID != uuid.Nil {
		exists, err := s.spaceRepo.Exists(ctx, filter.SpaceID)
		if err != nil {
			return nil, fmt.Errorf("failed to check space existence: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("space not found: %s", filter.SpaceID)
		}
	}

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

	if err := EnforceOwner(ctx, content.OwnerID); err != nil {
		return nil, err
	}

	// Idempotency check: if already in the target status, return early
	if content.Status == status {
		s.logger.Printf("Content source %s already in status %s, skipping update", id, status)
		return content, nil
	}

	// Validate status transition
	if err := content.Status.ValidateTransition(status); err != nil {
		s.logger.Printf("Invalid status transition for content source %s: %v", id, err)
	}

	// Update status
	err = s.contentRepo.UpdateStatus(ctx, id, status, processedBlobHash)
	if err != nil {
		return nil, fmt.Errorf("failed to update content source status: %w", err)
	}

	// Get updated content source
	content, err = s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated content source: %w", err)
	}

	return content, nil
}

// UpdateContentSource performs partial updates on title and keywords only.
func (s *ContentService) UpdateContentSource(ctx context.Context, id uuid.UUID, title *string, keywords *[]string) (*domain.ContentSource, error) {
	// Load current content
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get content source: %w", err)
	}

	// Enforce RLS
	if err := EnforceOwner(ctx, content.OwnerID); err != nil {
		return nil, err
	}

	// Apply partial updates
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		content.Title = trimmed
	}
	if keywords != nil {
		// allow empty slice to clear keywords
		content.Keywords = *keywords
	}

	if err := s.contentRepo.Update(ctx, content); err != nil {
		return nil, fmt.Errorf("failed to update content source: %w", err)
	}

	// Re-fetch to return fresh timestamps/state
	updated, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated content source: %w", err)
	}
	return updated, nil
}

// SpaceContentIntegrityReport represents the results of space-content validation
type SpaceContentIntegrityReport struct {
	TotalContentSources int               `json:"total_content_sources"`
	OrphanedCount       int               `json:"orphaned_count"`
	OrphanedContent     []uuid.UUID       `json:"orphaned_content_ids"`
	ValidatedSpaces     map[uuid.UUID]int `json:"validated_spaces"` // spaceID -> content count
}

// ValidateSpaceContentIntegrity checks for orphaned content sources and returns validation results
func (s *ContentService) ValidateSpaceContentIntegrity(ctx context.Context) (*SpaceContentIntegrityReport, error) {
	// Get all content sources
	allContent, err := s.contentRepo.List(ctx, domain.ContentSourceFilter{Limit: 10000}) // Large limit for validation
	if err != nil {
		return nil, fmt.Errorf("failed to list all content: %w", err)
	}

	report := &SpaceContentIntegrityReport{
		TotalContentSources: len(allContent),
		OrphanedContent:     []uuid.UUID{},
		ValidatedSpaces:     make(map[uuid.UUID]int),
	}

	// Check each content source's space
	for _, content := range allContent {
		exists, err := s.spaceRepo.Exists(ctx, content.SpaceID)
		if err != nil {
			s.logger.Printf("Error checking space %s for content %s: %v", content.SpaceID, content.ID, err)
			continue
		}

		if !exists {
			report.OrphanedContent = append(report.OrphanedContent, content.ID)
		} else {
			report.ValidatedSpaces[content.SpaceID]++
		}
	}

	report.OrphanedCount = len(report.OrphanedContent)
	return report, nil
}

// GenerateDownloadURL returns a short-lived pre-signed URL to download the original uploaded blob.
func (s *ContentService) GenerateDownloadURL(ctx context.Context, contentID uuid.UUID, expires time.Duration) (string, time.Time, error) {
	return s.GenerateDownloadURLWithKind(ctx, contentID, expires, "source")
}

// GenerateDownloadURLWithKind generates a presigned download URL for the requested kind.
func (s *ContentService) GenerateDownloadURLWithKind(ctx context.Context, contentID uuid.UUID, expires time.Duration, kind string) (string, time.Time, error) {
	if contentID == uuid.Nil {
		return "", time.Time{}, fmt.Errorf("content_id is required")
	}

	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get content source: %w", err)
	}

	// Enforce RLS: caller must be the owner
	if err := EnforceOwner(ctx, content.OwnerID); err != nil {
		return "", time.Time{}, err
	}

	bucket, err := s.resolveBucket(kind)
	if err != nil {
		return "", time.Time{}, err
	}

	if expires <= 0 {
		expires = 15 * time.Minute
	}

	filename := content.Source
	if strings.ToLower(strings.TrimSpace(kind)) == "processed" {
		filename = processedFilename(filename)
	}
	objectKey := buildObjectKey(content.OwnerID, content.SpaceID, content.ID, filename)

	url, err := s.storage.GeneratePresignedDownloadURL(bucket, objectKey, expires)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, time.Now().Add(expires), nil
}

// DeleteContentSource deletes the source object in R2 (best-effort) and removes the DB row.
func (s *ContentService) DeleteContentSource(ctx context.Context, id uuid.UUID) error {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get content source: %w", err)
	}

	// Enforce RLS
	if err := EnforceOwner(ctx, content.OwnerID); err != nil {
		return err
	}

	bucket, err := s.resolveBucket("source")
	if err != nil {
		return fmt.Errorf("failed to resolve source bucket: %w", err)
	}

	filename := strings.TrimSpace(content.Source)
	origKey := buildObjectKey(content.OwnerID, content.SpaceID, content.ID, filename)
	procKey := buildObjectKey(content.OwnerID, content.SpaceID, content.ID, processedFilename(filename))
	if err := s.storage.DeleteObject(bucket, origKey); err != nil {
		s.logger.Printf("WARN: failed to delete R2 object (bucket=%s key=%s): %v", bucket, origKey, err)
	}
	if err := s.storage.DeleteObject(bucket, procKey); err != nil {
		s.logger.Printf("WARN: failed to delete R2 object (bucket=%s key=%s): %v", bucket, procKey, err)
	}

	if err := s.contentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete content source from DB: %w", err)
	}

	return nil
}

// GetContentOwner returns the owner UUID for a content source (no RLS checks; for internal use)
func (s *ContentService) GetContentOwner(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get content source: %w", err)
	}
	return content.OwnerID, nil
}
