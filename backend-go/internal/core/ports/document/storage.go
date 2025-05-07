package document

import (
	"context"
	"io"

	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// StoragePort defines the interface for document storage operations
type StoragePort interface {
	// Upload stores document content and returns storage metadata
	// The content will be streamed from the provided reader
	Upload(ctx context.Context, content *document.DocumentContent, options *document.StorageOptions) (*document.StorageObject, error)

	// Download retrieves document content from storage
	// Will stream the content through the returned DocumentContent.Stream
	Download(ctx context.Context, objectID document.StorageObjectID) (*document.DocumentContent, error)

	// DownloadToWriter downloads object content directly to a writer
	// Useful for streaming content directly to HTTP response
	DownloadToWriter(ctx context.Context, objectID document.StorageObjectID, writer io.Writer) error

	// GetPresignedUploadURL generates a pre-signed URL for client-direct uploads
	// Allows web clients to upload directly to storage without routing through the API
	GetPresignedUploadURL(ctx context.Context, filename string, contentType string, options *document.StorageOptions) (*document.StorageUploadInfo, error)

	// GetPresignedDownloadURL generates a pre-signed URL for client-direct downloads
	// Allows web clients to download directly from storage without routing through the API
	GetPresignedDownloadURL(ctx context.Context, objectID document.StorageObjectID, filename string, expiresInSeconds int) (*document.StorageDownloadInfo, error)

	// Delete removes an object from storage
	Delete(ctx context.Context, objectID document.StorageObjectID) error

	// Copy duplicates an object to a new location/key
	Copy(ctx context.Context, sourceID document.StorageObjectID, destOptions *document.StorageOptions) (*document.StorageObject, error)

	// GetObjectMetadata retrieves metadata without downloading content
	GetObjectMetadata(ctx context.Context, objectID document.StorageObjectID) (*document.StorageObject, error)

	// ListObjectsByPrefix lists objects with a specific prefix (path)
	// Useful for cleaning up related files or listing document attachments
	ListObjectsByPrefix(ctx context.Context, prefix string, maxItems int) ([]*document.StorageObject, error)

	// ConvertFileFormat attempts to convert a file from one format to another
	// This can be used for preprocessing unsupported file types
	// Returns the ID of the converted file
	ConvertFileFormat(ctx context.Context, sourceID document.StorageObjectID, targetFormat string) (document.StorageObjectID, error)
}
