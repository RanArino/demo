package models

import (
	"time"
)

// ProcessingStatus represents the current state of document processing
type ProcessingStatus string

const (
	StatusUploaded   ProcessingStatus = "UPLOADED"
	StatusProcessing ProcessingStatus = "PROCESSING"
	StatusCompleted  ProcessingStatus = "COMPLETED"
	StatusFailed     ProcessingStatus = "FAILED"
)

// Document represents an uploaded document and its processing state
type Document struct {
	ID          string           `json:"id"`
	Filename    string           `json:"filename"`
	Status      ProcessingStatus `json:"status"`
	SummaryID   *string          `json:"summary_id,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	ProcessedAt *time.Time       `json:"processed_at,omitempty"`
	Error       *string          `json:"error,omitempty"`
	Keywords    []string         `json:"keywords,omitempty"`
}

// Chunk represents a segment of text from a document
type Chunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Index      int       `json:"index"`
	Text       string    `json:"text"`
	TokenCount int       `json:"token_count"`
	Embedding  []float32 `json:"embedding,omitempty"`
	Keywords   []string  `json:"keywords,omitempty"`
	// Visualization data (to be used later)
	Coord2D    *[2]float32 `json:"coord_2d,omitempty"`
	Coord3D    *[3]float32 `json:"coord_3d,omitempty"`
	ClusterIDs []int       `json:"cluster_ids,omitempty"`
}

// Summary represents an AI-generated summary of a document
type Summary struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Text       string    `json:"text"`
	Embedding  []float32 `json:"embedding,omitempty"`
	// Visualization data (to be used later)
	Coord2D   *[2]float32 `json:"coord_2d,omitempty"`
	Coord3D   *[3]float32 `json:"coord_3d,omitempty"`
	ClusterID *int        `json:"cluster_id,omitempty"`
}

// Validate checks if the Document struct has all required fields
func (d *Document) Validate() error {
	if d.ID == "" {
		return ErrEmptyID
	}
	if d.Filename == "" {
		return ErrEmptyFilename
	}
	return nil
}

// Validate checks if the Chunk struct has all required fields
func (c *Chunk) Validate() error {
	if c.ID == "" {
		return ErrEmptyID
	}
	if c.DocumentID == "" {
		return ErrEmptyDocumentID
	}
	if c.Text == "" {
		return ErrEmptyText
	}
	return nil
}

// Validate checks if the Summary struct has all required fields
func (s *Summary) Validate() error {
	if s.ID == "" {
		return ErrEmptyID
	}
	if s.DocumentID == "" {
		return ErrEmptyDocumentID
	}
	if s.Text == "" {
		return ErrEmptyText
	}
	return nil
}
