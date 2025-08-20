package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"demo/ms_knowledge/internal/domain"
	"demo/ms_knowledge/internal/events"

	"github.com/google/uuid"
)

// DocumentWorkflowIntegrationTest tests the complete end-to-end document workflow
// from upload creation through processing completion
func TestDocumentWorkflowIntegration_SuccessfulFlow(t *testing.T) {
	// Setup test environment
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}

	// Mock event producer that captures events
	eventCapture := &mockEventProducer{events: make([]mockEvent, 0)}

	// Services
	contentSvc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, eventCapture, "test-bucket", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
	spaceSvc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()

	// Step 1: Create a space
	space, err := spaceSvc.CreateSpace(ctx, "Test Space", "Integration test space", uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to create space: %v", err)
	}

	// Step 2: Create upload URL
	content, uploadURL, err := contentSvc.CreateUploadURL(ctx, space.ID, "test-document.pdf", "application/pdf", 1024, "Test Document")
	if err != nil {
		t.Fatalf("Failed to create upload URL: %v", err)
	}

	// Validate initial state
	if content.Status != domain.ContentStatusUploading {
		t.Errorf("Expected status UPLOADING, got %s", content.Status)
	}
	if uploadURL == "" {
		t.Error("Expected upload URL to be generated")
	}

	// Step 3: Confirm upload (simulates client completing upload)
	originalHash := "original-blob-hash-123"
	confirmedContent, err := contentSvc.ConfirmUpload(ctx, content.ID, originalHash)
	if err != nil {
		t.Fatalf("Failed to confirm upload: %v", err)
	}

	// Validate upload confirmation
	if confirmedContent.Status != domain.ContentStatusUploaded {
		t.Errorf("Expected status UPLOADED, got %s", confirmedContent.Status)
	}
	if confirmedContent.OriginalBlobHash != originalHash {
		t.Errorf("Expected original hash %s, got %s", originalHash, confirmedContent.OriginalBlobHash)
	}

	// Verify document.uploaded event was produced
	if len(eventCapture.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(eventCapture.events))
	}
	uploadedEvent := eventCapture.events[0]
	if uploadedEvent.topic != events.TopicDocumentUploaded {
		t.Errorf("Expected topic %s, got %s", events.TopicDocumentUploaded, uploadedEvent.topic)
	}

	// Step 4: Simulate document processing (ms_document_process would do this)
	// First, transition to PROCESSING status
	processingContent, err := contentSvc.UpdateContentSourceStatus(ctx, content.ID, domain.ContentStatusProcessing, "", "")
	if err != nil {
		t.Fatalf("Failed to update to processing status: %v", err)
	}

	if processingContent.Status != domain.ContentStatusProcessing {
		t.Errorf("Expected status PROCESSING, got %s", processingContent.Status)
	}

	// Step 5: Complete processing with successful result
	processedHash := "processed-blob-hash-456"
	processedContent, err := contentSvc.UpdateContentSourceStatus(ctx, content.ID, domain.ContentStatusProcessed, processedHash, "")
	if err != nil {
		t.Fatalf("Failed to update to processed status: %v", err)
	}

	// Validate final state
	if processedContent.Status != domain.ContentStatusProcessed {
		t.Errorf("Expected status PROCESSED, got %s", processedContent.Status)
	}
	if processedContent.ProcessedBlobHash == nil || *processedContent.ProcessedBlobHash != processedHash {
		t.Errorf("Expected processed hash %s, got %v", processedHash, processedContent.ProcessedBlobHash)
	}

	// Step 6: Verify space statistics are updated
	spaceWithStats, err := spaceSvc.GetSpace(ctx, space.ID)
	if err != nil {
		t.Fatalf("Failed to get space with stats: %v", err)
	}

	// Content count should be 1
	if spaceWithStats.Stats.ContentCount != 1 {
		t.Errorf("Expected content count 1, got %d", spaceWithStats.Stats.ContentCount)
	}
}

// TestDocumentWorkflowIntegration_FailureScenarios tests error handling in the workflow
func TestDocumentWorkflowIntegration_FailureScenarios(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	eventCapture := &mockEventProducer{events: make([]mockEvent, 0)}

	contentSvc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, eventCapture, "test-bucket", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
	spaceSvc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()

	// Create space and content
	space, _ := spaceSvc.CreateSpace(ctx, "Test Space", "Test space", uuid.New().String())
	content, _, _ := contentSvc.CreateUploadURL(ctx, space.ID, "test-doc.pdf", "application/pdf", 1024, "Test Doc")
	contentSvc.ConfirmUpload(ctx, content.ID, "original-hash")

	// Test 1: Processing failure
	contentSvc.UpdateContentSourceStatus(ctx, content.ID, domain.ContentStatusProcessing, "", "")

	// Simulate processing failure
	failedContent, err := contentSvc.UpdateContentSourceStatus(ctx, content.ID, domain.ContentStatusFailed, "", "Processing failed due to invalid PDF format")
	if err != nil {
		t.Fatalf("Failed to update to failed status: %v", err)
	}

	if failedContent.Status != domain.ContentStatusFailed {
		t.Errorf("Expected status FAILED, got %s", failedContent.Status)
	}

	// Test 2: Invalid status transitions (should be logged but allowed for now)
	// Try to go from FAILED directly to PROCESSED (invalid transition)
	_, err = contentSvc.UpdateContentSourceStatus(ctx, content.ID, domain.ContentStatusProcessed, "hash", "")
	// This should succeed for now (as per implementation) but log a warning
	if err != nil {
		t.Fatalf("Status transition validation should allow invalid transitions for now: %v", err)
	}
}

// TestDocumentWorkflowIntegration_EventHandling tests the event handler implementation
func TestDocumentWorkflowIntegration_EventHandling(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}

	contentSvc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, "test-bucket", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
	spaceSvc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()

	// Setup test data
	space, _ := spaceSvc.CreateSpace(ctx, "Test Space", "Test space", uuid.New().String())
	content, _, _ := contentSvc.CreateUploadURL(ctx, space.ID, "test-doc.pdf", "application/pdf", 1024, "Test Doc")
	contentSvc.ConfirmUpload(ctx, content.ID, "original-hash")
	contentSvc.UpdateContentSourceStatus(ctx, content.ID, domain.ContentStatusProcessing, "", "")

	// Create event handler
	handler := &DocumentProcessedHandler{
		contentSvc: contentSvc,
		logger:     log.New(os.Stdout, "[HANDLER] ", log.LstdFlags),
	}

	// Test successful processing event
	successEvent := events.DocumentProcessedEvent{
		ContentSourceID:   content.ID,
		ProcessedBlobHash: stringPtr("processed-hash-123"),
		Status:            events.ProcessStatusProcessed,
	}

	err := handler.HandleDocumentProcessed(ctx, successEvent)
	if err != nil {
		t.Fatalf("Failed to handle successful processing event: %v", err)
	}

	// Verify content was updated
	updatedContent, _ := contentSvc.GetContentSource(ctx, content.ID)
	if updatedContent.Status != domain.ContentStatusProcessed {
		t.Errorf("Expected status PROCESSED, got %s", updatedContent.Status)
	}

	// Test failure event
	content2, _, _ := contentSvc.CreateUploadURL(ctx, space.ID, "test-doc2.pdf", "application/pdf", 1024, "Test Doc 2")
	contentSvc.ConfirmUpload(ctx, content2.ID, "original-hash-2")
	contentSvc.UpdateContentSourceStatus(ctx, content2.ID, domain.ContentStatusProcessing, "", "")

	failureEvent := events.DocumentProcessedEvent{
		ContentSourceID: content2.ID,
		Status:          events.ProcessStatusFailed,
		ErrorMessage:    "Failed to process document",
	}

	err = handler.HandleDocumentProcessed(ctx, failureEvent)
	if err != nil {
		t.Fatalf("Failed to handle failure processing event: %v", err)
	}

	// Verify content was updated to failed
	updatedContent2, _ := contentSvc.GetContentSource(ctx, content2.ID)
	if updatedContent2.Status != domain.ContentStatusFailed {
		t.Errorf("Expected status FAILED, got %s", updatedContent2.Status)
	}
}

// TestDocumentWorkflowIntegration_ConcurrentOperations tests concurrent operations on the same content
func TestDocumentWorkflowIntegration_ConcurrentOperations(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	eventCapture := &mockEventProducer{events: make([]mockEvent, 0)}

	contentSvc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, eventCapture, "test-bucket", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
	spaceSvc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()

	// Setup
	space, _ := spaceSvc.CreateSpace(ctx, "Test Space", "Test space", uuid.New().String())
	content, _, _ := contentSvc.CreateUploadURL(ctx, space.ID, "test-doc.pdf", "application/pdf", 1024, "Test Doc")
	contentSvc.ConfirmUpload(ctx, content.ID, "original-hash")

	// Test concurrent status updates (simulates race conditions)
	var wg sync.WaitGroup
	results := make([]error, 10)

	// Launch 10 concurrent goroutines trying to update status
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := contentSvc.UpdateContentSourceStatus(ctx, content.ID, domain.ContentStatusProcessing, "", "")
			results[index] = err
		}(i)
	}

	wg.Wait()

	// All operations should succeed (idempotency)
	for i, err := range results {
		if err != nil {
			t.Errorf("Concurrent operation %d failed: %v", i, err)
		}
	}

	// Final status should be PROCESSING
	finalContent, _ := contentSvc.GetContentSource(ctx, content.ID)
	if finalContent.Status != domain.ContentStatusProcessing {
		t.Errorf("Expected final status PROCESSING, got %s", finalContent.Status)
	}
}

// TestDocumentWorkflowIntegration_SpaceContentIntegrity tests data integrity validation
func TestDocumentWorkflowIntegration_SpaceContentIntegrity(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}

	contentSvc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, "test-bucket", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
	spaceSvc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()

	// Create spaces and content
	space1, _ := spaceSvc.CreateSpace(ctx, "Space 1", "Space 1", uuid.New().String())
	space2, _ := spaceSvc.CreateSpace(ctx, "Space 2", "Space 2", uuid.New().String())

	// Add content to both spaces
	_, _, _ = contentSvc.CreateUploadURL(ctx, space1.ID, "doc1.pdf", "application/pdf", 1024, "Doc 1")
	_, _, _ = contentSvc.CreateUploadURL(ctx, space2.ID, "doc2.pdf", "application/pdf", 1024, "Doc 2")

	// Create orphaned content (simulate space deletion)
	orphanedSpaceID := uuid.New()
	orphanedContent := &domain.ContentSource{
		ID:        uuid.New(),
		SpaceID:   orphanedSpaceID,
		Title:     "Orphaned Doc",
		MediaType: "application/pdf",
		Source:    "orphaned.pdf",
		Status:    domain.ContentStatusUploaded,
	}
	contentRepo.Create(ctx, orphanedContent)

	// Run integrity check
	report, err := contentSvc.ValidateSpaceContentIntegrity(ctx)
	if err != nil {
		t.Fatalf("Failed to validate integrity: %v", err)
	}

	// Validate report
	if report.TotalContentSources != 3 {
		t.Errorf("Expected 3 total content sources, got %d", report.TotalContentSources)
	}

	if report.OrphanedCount != 1 {
		t.Errorf("Expected 1 orphaned content, got %d", report.OrphanedCount)
	}

	if len(report.OrphanedContent) != 1 || report.OrphanedContent[0] != orphanedContent.ID {
		t.Errorf("Expected orphaned content ID %s, got %v", orphanedContent.ID, report.OrphanedContent)
	}

	// Should have 2 validated spaces
	if len(report.ValidatedSpaces) != 2 {
		t.Errorf("Expected 2 validated spaces, got %d", len(report.ValidatedSpaces))
	}

	// Each space should have 1 content
	if count, ok := report.ValidatedSpaces[space1.ID]; !ok || count != 1 {
		t.Errorf("Expected space1 to have 1 content, got %d", count)
	}
	if count, ok := report.ValidatedSpaces[space2.ID]; !ok || count != 1 {
		t.Errorf("Expected space2 to have 1 content, got %d", count)
	}
}

// Mock event producer for testing
type mockEventProducer struct {
	mu     sync.Mutex
	events []mockEvent
}

type mockEvent struct {
	topic string
	key   string
	data  interface{}
}

func (m *mockEventProducer) ProduceJSON(ctx context.Context, topic string, key string, v any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = append(m.events, mockEvent{
		topic: topic,
		key:   key,
		data:  v,
	})
	return nil
}

func (m *mockEventProducer) Close() {}

// DocumentProcessedHandler implements the event handler for testing
type DocumentProcessedHandler struct {
	contentSvc *ContentService
	logger     *log.Logger
}

func (h *DocumentProcessedHandler) HandleDocumentProcessed(ctx context.Context, event events.DocumentProcessedEvent) error {
	h.logger.Printf("Handling document processed event: %+v", event)

	// Convert event status to domain status
	var status domain.ContentStatus
	switch event.Status {
	case events.ProcessStatusProcessed:
		status = domain.ContentStatusProcessed
	case events.ProcessStatusFailed:
		status = domain.ContentStatusFailed
	case events.ProcessStatusProcessing:
		status = domain.ContentStatusProcessing
	default:
		return fmt.Errorf("unknown process status: %s", event.Status)
	}

	// Extract processed blob hash
	processedHash := ""
	if event.ProcessedBlobHash != nil {
		processedHash = *event.ProcessedBlobHash
	}

	// Update content source status
	_, err := h.contentSvc.UpdateContentSourceStatus(ctx, event.ContentSourceID, status, processedHash, event.ErrorMessage)
	if err != nil {
		return fmt.Errorf("failed to update content source status: %w", err)
	}

	h.logger.Printf("Successfully updated content source %s to status %s", event.ContentSourceID, status)
	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
