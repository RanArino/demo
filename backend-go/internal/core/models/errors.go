package models

import (
	"errors"
	"fmt"
)

// Core validation errors
var (
	ErrEmptyID         = errors.New("id cannot be empty")
	ErrEmptyFilename   = errors.New("filename cannot be empty")
	ErrEmptyDocumentID = errors.New("document_id cannot be empty")
	ErrEmptyText       = errors.New("text content cannot be empty")
)

// Repository errors
var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrChunkNotFound    = errors.New("chunk not found")
	ErrSummaryNotFound  = errors.New("summary not found")
)

// Vector store errors
var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrCollectionExists   = errors.New("collection already exists")
	ErrInvalidVectorSize  = errors.New("invalid vector size")
	ErrPointNotFound      = errors.New("point not found in collection")
	ErrInvalidEmbedding   = errors.New("invalid embedding vector")
)

// LLM service errors
var (
	ErrEmbeddingGeneration = errors.New("failed to generate embedding")
	ErrSummaryGeneration   = errors.New("failed to generate summary")
	ErrTokenLimitExceeded  = errors.New("text exceeds token limit")
)

// NewErrInvalidVectorSize creates a new error for invalid vector size
func NewErrInvalidVectorSize(expected, got uint64) error {
	return fmt.Errorf("%w: expected %d, got %d", ErrInvalidVectorSize, expected, got)
}

// NewErrEmbeddingGeneration creates a new error for embedding generation failure
func NewErrEmbeddingGeneration(cause error) error {
	return fmt.Errorf("%w: %v", ErrEmbeddingGeneration, cause)
}

// NewErrSummaryGeneration creates a new error for summary generation failure
func NewErrSummaryGeneration(cause error) error {
	return fmt.Errorf("%w: %v", ErrSummaryGeneration, cause)
}
