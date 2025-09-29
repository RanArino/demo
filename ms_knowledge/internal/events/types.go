package events

import (
	"strings"

	"github.com/google/uuid"
)

// Topic names used by the system. Keeping them centralized avoids typos.
const (
	TopicDocumentUploaded  = "document.uploaded"
	TopicDocumentProcessed = "document.processed"
)

// DocumentUploadedEvent is produced by ms_knowledge when a client confirms an upload.
type DocumentUploadedEvent struct {
	ContentSourceID   uuid.UUID `json:"content_source_id"`
	OriginalBlobHash  string    `json:"original_blob_hash"`
	SpaceID           uuid.UUID `json:"space_id"`
	OriginalObjectKey string    `json:"original_object_key"`
	Title             string    `json:"title"`
}

// ProcessStatus is the processing outcome in the processed event.
type ProcessStatus string

const (
	ProcessStatusProcessed  ProcessStatus = "PROCESSED"
	ProcessStatusFailed     ProcessStatus = "FAILED"
	ProcessStatusPending    ProcessStatus = "PENDING"
	ProcessStatusProcessing ProcessStatus = "PROCESSING"
)

// DocumentProcessedEvent is produced by ms_document_process after processing.
type DocumentProcessedEvent struct {
	ContentSourceID    uuid.UUID     `json:"content_source_id"`
	SpaceID            uuid.UUID     `json:"space_id"`
	ProcessedBlobHash  *string       `json:"processed_blob_hash,omitempty"`
	ProcessedObjectKey *string       `json:"processed_object_key,omitempty"`
	Status             ProcessStatus `json:"status"`
	ErrorMessage       string        `json:"error_message,omitempty"`
	Keywords           []string      `json:"keywords,omitempty"`
	Summary            string        `json:"summary,omitempty"`
	Title              string        `json:"title,omitempty"`
}

func (e DocumentProcessedEvent) SummaryPtr() *string {
	if strings.TrimSpace(e.Summary) == "" {
		return nil
	}
	return &e.Summary
}

func (e DocumentProcessedEvent) KeywordsPtr() *[]string {
	if len(e.Keywords) == 0 {
		return nil
	}
	return &e.Keywords
}

func (e DocumentProcessedEvent) TitlePtr() *string {
	trimmed := strings.TrimSpace(e.Title)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
