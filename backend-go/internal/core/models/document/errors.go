package document

import (
	"errors"
	"fmt"
)

// Common errors for document operations
var (
	ErrDocumentNotFound      = errors.New("document not found")
	ErrDocumentAlreadyExists = errors.New("document already exists")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrInvalidDocumentID     = errors.New("invalid document ID")
	ErrInvalidDocumentType   = errors.New("invalid document type")
	ErrDocumentProcessing    = errors.New("document is currently being processed")
	ErrContentNotProcessed   = errors.New("document content has not been processed")
	ErrStorageAccessDenied   = errors.New("access to storage denied")
)

// StorageError represents an error from a storage service
type StorageError struct {
	Operation string
	Key       string
	Provider  string
	Err       error
}

// Error returns the error message
func (e *StorageError) Error() string {
	return fmt.Sprintf("storage error during %s operation on %s (provider: %s): %v",
		e.Operation, e.Key, e.Provider, e.Err)
}

// Unwrap returns the underlying error
func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError creates a new storage error
func NewStorageError(err error, operation, key, provider string) error {
	return &StorageError{
		Operation: operation,
		Key:       key,
		Provider:  provider,
		Err:       err,
	}
}

// Document Service specific errors
var (
	// Storage errors
	ErrStorageObjectNotFound = errors.New("storage object not found")
	ErrStorageBucketNotFound = errors.New("storage bucket not found")
	ErrStorageUploadFailed   = errors.New("failed to upload object to storage")
	ErrStorageDownloadFailed = errors.New("failed to download object from storage")
	ErrStorageDeleteFailed   = errors.New("failed to delete object from storage")

	// Parsing errors
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrFileTooLarge        = errors.New("file is too large")
	ErrParsingFailed       = errors.New("failed to parse document")
	ErrNoParsingStrategy   = errors.New("no suitable parsing strategy found")

	// Queue errors
	ErrQueueSendFailed                = errors.New("failed to send message to queue")
	ErrQueueReceiveFailed             = errors.New("failed to receive messages from queue")
	ErrQueueDeleteFailed              = errors.New("failed to delete message from queue")
	ErrQueueVisibilityExtensionFailed = errors.New("failed to extend message visibility timeout")

	// External provider errors
	ErrExternalProviderAuth         = errors.New("external provider authentication failed")
	ErrExternalProviderFetch        = errors.New("external provider fetch failed")
	ErrExternalProviderNotSupported = errors.New("external provider not supported")
)

// DocumentServiceError represents a custom error in the document service
type DocumentServiceError struct {
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface for DocumentServiceError
func (e *DocumentServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the cause of the error
func (e *DocumentServiceError) Unwrap() error {
	return e.Cause
}

// NewDocumentServiceError creates a new document service error
func NewDocumentServiceError(code string, message string, cause error) *DocumentServiceError {
	return &DocumentServiceError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewParsingError creates a new document parsing error
func NewParsingError(baseErr error, documentID DocumentID, details string) error {
	return NewDocumentServiceError(
		"PARSING_ERROR",
		fmt.Sprintf("Error parsing document %s: %s", documentID, details),
		baseErr,
	)
}

// NewQueueError creates a new queue operation error
func NewQueueError(baseErr error, operation string, messageID string) error {
	return NewDocumentServiceError(
		"QUEUE_ERROR",
		fmt.Sprintf("Queue error during %s operation on message %s", operation, messageID),
		baseErr,
	)
}
