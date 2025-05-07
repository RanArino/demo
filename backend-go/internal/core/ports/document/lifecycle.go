package document

import (
	"context"
	"io"

	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// ImportSource represents the source of an external document import
type ImportSource string

const (
	ImportSourceGoogleDrive ImportSource = "GOOGLE_DRIVE"
	ImportSourceOneDrive    ImportSource = "ONEDRIVE"
	ImportSourceDropbox     ImportSource = "DROPBOX"
	ImportSourceURL         ImportSource = "URL"
)

// ImportRequest contains information for importing a document from an external source
type ImportRequest struct {
	Source      ImportSource        // Source system (e.g., Google Drive)
	ExternalID  string              // ID in the external system
	AccessToken string              // OAuth token for the external system
	Metadata    map[string]string   // Additional metadata about the document
	SpaceID     string              // Space to import into
	Options     *FileParsingOptions // Options for parsing the document
}

// DocumentLifecycleManagerPort orchestrates the complete document lifecycle
type DocumentLifecycleManagerPort interface {
	// HandleFileUpload processes a direct file upload
	// Verifies user permission to upload to the space (Editor+)
	// Returns document metadata with initial ID and status
	HandleFileUpload(ctx context.Context, userID string, spaceID string, file *document.DocumentContent) (*document.DocumentMetadata, error)

	// HandleExternalImport processes a document import from an external source
	// Verifies user permission to import to the space (Editor+)
	HandleExternalImport(ctx context.Context, userID string, importRequest *ImportRequest) (*document.DocumentMetadata, error)

	// GetDocumentDetails retrieves detailed metadata about a document
	// Verifies user has permission to view the document (Viewer+)
	GetDocumentDetails(ctx context.Context, userID string, docID document.DocumentID) (*document.DocumentMetadata, error)

	// GetDocumentContent retrieves the processed content of a document
	// Verifies user has permission to view the document (Viewer+)
	GetDocumentContent(ctx context.Context, userID string, docID document.DocumentID) (*document.StructuredDocumentOutput, error)

	// GetMarkdownContent retrieves just the markdown content of a document
	// Verifies user has permission to view the document (Viewer+)
	GetMarkdownContent(ctx context.Context, userID string, docID document.DocumentID) (string, error)

	// StreamOriginalContent streams the original document file to the provided writer
	// Verifies user has permission to view the document (Viewer+)
	StreamOriginalContent(ctx context.Context, userID string, docID document.DocumentID, writer io.Writer) error

	// GetDownloadURL generates a pre-signed URL for downloading the original document
	// Verifies user has permission to view the document (Viewer+)
	GetDownloadURL(ctx context.Context, userID string, docID document.DocumentID) (*document.StorageDownloadInfo, error)

	// GetPresignedUploadURL generates a pre-signed URL for direct upload
	// This is a convenience method for client-direct uploads
	// Verifies user has permission to upload to the space (Editor+)
	GetPresignedUploadURL(ctx context.Context, userID string, spaceID string, filename string, contentType string) (*document.StorageUploadInfo, error)

	// DeleteDocument deletes a document
	// Soft-delete by default, marking as deleted in the repository
	// Verifies user has permission to delete (Editor+)
	DeleteDocument(ctx context.Context, userID string, docID document.DocumentID, permanent bool) error

	// ReprocessDocument forces reprocessing of a document
	// Useful if the original processing failed or needs to be updated
	// Verifies user has permission to edit the document (Editor+)
	ReprocessDocument(ctx context.Context, userID string, docID document.DocumentID, options *FileParsingOptions) (*document.DocumentMetadata, error)

	// ListUserDocuments lists documents the user has access to
	// Automatically filters based on user permissions across spaces
	ListUserDocuments(ctx context.Context, userID string, spaceID *string, offset, limit int) ([]*document.DocumentMetadata, int, error)

	// CanUserPerformAction checks if a user can perform an action on a document
	// Used for permission checking in the API layer
	CanUserPerformAction(ctx context.Context, userID string, docID document.DocumentID, action string) (bool, error)
}
