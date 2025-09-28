package events

import "github.com/google/uuid"

type ProcessStatus string

const (
	ProcessStatusProcessed  ProcessStatus = "PROCESSED"
	ProcessStatusFailed     ProcessStatus = "FAILED"
	ProcessStatusPending    ProcessStatus = "PENDING"
	ProcessStatusProcessing ProcessStatus = "PROCESSING"
)

// DocumentProcessedEvent mirrors the event produced by ms_document_process
// and consumed by ms_canvas.
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
