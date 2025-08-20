package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestConsumerSimple_RetryMechanism tests the retry logic with simpler mocking
func TestConsumerSimple_RetryMechanism(t *testing.T) {
	handler := &simpleProcessedHandler{
		processedEvents: make([]DocumentProcessedEvent, 0),
		shouldError:     true,
		errorCount:      0,
		maxErrors:       2, // Fail first 2 attempts, succeed on 3rd
	}
	
	ctx := context.Background()
	
	// Create test event
	testEvent := DocumentProcessedEvent{
		ContentSourceID:   uuid.New(),
		ProcessedBlobHash: stringPtr("test-hash"),
		Status:            ProcessStatusProcessed,
	}
	
	// Simulate the retry logic directly
	maxRetries := 3
	var finalErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := handler.HandleDocumentProcessed(ctx, testEvent); err != nil {
			if attempt == maxRetries {
				finalErr = err
				break
			}
			// Wait a bit between retries like the real implementation
			time.Sleep(time.Millisecond * 10)
			continue
		} else {
			break // Success
		}
	}
	
	// Should succeed eventually
	if finalErr != nil {
		t.Errorf("Expected eventual success, but got final error: %v", finalErr)
	}
	
	// Verify retry attempts were made (should be 3 attempts total: initial + 2 retries)
	if handler.errorCount != 2 {
		t.Errorf("Expected 2 error attempts before success, got %d", handler.errorCount)
	}
	
	// Verify the event was eventually processed successfully
	if len(handler.processedEvents) != 1 {
		t.Errorf("Expected 1 processed event after retries, got %d", len(handler.processedEvents))
	}
}

// TestConsumerSimple_InvalidJSON tests JSON unmarshaling behavior
func TestConsumerSimple_InvalidJSON(t *testing.T) {
	handler := &simpleProcessedHandler{
		processedEvents: make([]DocumentProcessedEvent, 0),
		shouldError:     false,
	}
	
	// Try to unmarshal invalid JSON
	invalidJSON := []byte("invalid json {{{")
	var payload DocumentProcessedEvent
	err := json.Unmarshal(invalidJSON, &payload)
	
	// Should fail to unmarshal
	if err == nil {
		t.Error("Expected JSON unmarshal error for invalid JSON")
	}
	
	// Handler should not have been called
	if len(handler.processedEvents) != 0 {
		t.Errorf("Expected 0 processed events for invalid JSON, got %d", len(handler.processedEvents))
	}
}

// TestConsumerSimple_EventTypes tests different event statuses
func TestConsumerSimple_EventTypes(t *testing.T) {
	handler := &simpleProcessedHandler{
		processedEvents: make([]DocumentProcessedEvent, 0),
		shouldError:     false,
	}
	
	ctx := context.Background()
	
	// Test different event types
	testEvents := []DocumentProcessedEvent{
		{
			ContentSourceID:   uuid.New(),
			ProcessedBlobHash: stringPtr("processed-hash-1"),
			Status:            ProcessStatusProcessed,
		},
		{
			ContentSourceID: uuid.New(),
			Status:          ProcessStatusFailed,
			ErrorMessage:    "Processing failed",
		},
		{
			ContentSourceID: uuid.New(),
			Status:          ProcessStatusProcessing,
		},
	}
	
	// Process all events
	for _, event := range testEvents {
		err := handler.HandleDocumentProcessed(ctx, event)
		if err != nil {
			t.Errorf("Failed to handle event: %v", err)
		}
	}
	
	// Verify all events were processed
	if len(handler.processedEvents) != len(testEvents) {
		t.Errorf("Expected %d processed events, got %d", len(testEvents), len(handler.processedEvents))
	}
	
	// Verify event content
	for i, expected := range testEvents {
		actual := handler.processedEvents[i]
		if actual.ContentSourceID != expected.ContentSourceID {
			t.Errorf("Event %d: Expected ContentSourceID %s, got %s", i, expected.ContentSourceID, actual.ContentSourceID)
		}
		if actual.Status != expected.Status {
			t.Errorf("Event %d: Expected Status %s, got %s", i, expected.Status, actual.Status)
		}
		if expected.ProcessedBlobHash != nil && actual.ProcessedBlobHash != nil {
			if *actual.ProcessedBlobHash != *expected.ProcessedBlobHash {
				t.Errorf("Event %d: Expected ProcessedBlobHash %s, got %s", i, *expected.ProcessedBlobHash, *actual.ProcessedBlobHash)
			}
		}
	}
}

// Simple mock handler for testing
type simpleProcessedHandler struct {
	mu              sync.Mutex
	processedEvents []DocumentProcessedEvent
	shouldError     bool
	errorCount      int
	maxErrors       int
}

func (m *simpleProcessedHandler) HandleDocumentProcessed(ctx context.Context, event DocumentProcessedEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldError && m.errorCount < m.maxErrors {
		m.errorCount++
		return fmt.Errorf("simulated handler error (attempt %d)", m.errorCount)
	}
	
	m.processedEvents = append(m.processedEvents, event)
	return nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

