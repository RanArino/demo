package events

import "github.com/google/uuid"

// DocumentProcessedEvent mirrors the event produced by ms_document_process
// and consumed by ms_canvas.
type DocumentProcessedEvent struct {
	ContentSourceID   uuid.UUID `json:"content_source_id"`
	SpaceID           uuid.UUID `json:"space_id"`
	ProcessedBlobHash *string   `json:"processed_blob_hash"`
	Status            string    `json:"status"`
	ErrorMessage      *string   `json:"error_message"`
}
