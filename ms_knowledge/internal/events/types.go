package events

import (
	"github.com/google/uuid"
)

// Topic names used by the system. Keeping them centralized avoids typos.
const (
	TopicDocumentUploaded  = "document.uploaded"
	TopicDocumentProcessed = "document.processed"
)

// DocumentUploadedEvent is produced by ms_knowledge when a client confirms an upload.
type DocumentUploadedEvent struct {
	ContentSourceID  uuid.UUID `json:"content_source_id"`
	OriginalBlobHash string    `json:"original_blob_hash"`
	SpaceID          uuid.UUID `json:"space_id"`
}

// ProcessStatus is the processing outcome in the processed event.
type ProcessStatus string

const (
	ProcessStatusProcessed ProcessStatus = "PROCESSED"
	ProcessStatusFailed    ProcessStatus = "FAILED"
)

// DocumentProcessedEvent is produced by ms_document_process after processing.
type DocumentProcessedEvent struct {
	ContentSourceID   uuid.UUID     `json:"content_source_id"`
	ProcessedBlobHash *string       `json:"processed_blob_hash,omitempty"`
	Status            ProcessStatus `json:"status"`
	ErrorMessage      string        `json:"error_message,omitempty"`
}
