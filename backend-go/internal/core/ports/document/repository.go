package document

import (
	"context"

	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// DocumentRepositoryPort defines the data access layer for document operations
type DocumentRepositoryPort interface {
	// CreateDocumentMetadata creates a new document metadata entry
	// Returns the document ID and error if any
	CreateDocumentMetadata(ctx context.Context, metadata *document.DocumentMetadata, userID string) (document.DocumentID, error)

	// GetDocumentMetadata retrieves document metadata by ID
	// Will verify if the requesting user has appropriate permissions to view the document
	GetDocumentMetadata(ctx context.Context, docID document.DocumentID, userID string) (*document.DocumentMetadata, error)

	// UpdateDocumentStatus updates the processing status of a document
	// Internal service operation - no user permission check needed as this is called by the system
	UpdateDocumentStatus(ctx context.Context, docID document.DocumentID, status document.DocumentStatus) error

	// StoreProcessedContent saves the structured document output after processing
	// Internal service operation
	StoreProcessedContent(ctx context.Context, docID document.DocumentID, structuredOutput *document.StructuredDocumentOutput) error

	// GetProcessedContent retrieves the full structured document output
	// Will verify if the requesting user has appropriate permissions
	GetProcessedContent(ctx context.Context, docID document.DocumentID, userID string) (*document.StructuredDocumentOutput, error)

	// GetProcessedMarkdown retrieves just the markdown content of a processed document
	// Will verify if the requesting user has appropriate permissions
	GetProcessedMarkdown(ctx context.Context, docID document.DocumentID, userID string) (string, error)

	// ListUserDocuments lists documents the user has access to with pagination
	// Will filter documents based on the user's permissions across spaces
	ListUserDocuments(ctx context.Context, userID string, spaceID *string, offset, limit int) ([]*document.DocumentMetadata, int, error)

	// DeleteDocument marks a document as deleted (soft delete)
	// Will verify if the requesting user has appropriate permissions (Editor+)
	DeleteDocument(ctx context.Context, docID document.DocumentID, userID string) error

	// PermanentlyDeleteDocument physically removes the document and its data
	// Will verify if the requesting user has appropriate permissions (Owner/Admin only)
	PermanentlyDeleteDocument(ctx context.Context, docID document.DocumentID, userID string) error

	// GetUserRoleForDocument retrieves the user's role for a specific document
	// This is used for permission checks in other services
	GetUserRoleForDocument(ctx context.Context, docID document.DocumentID, userID string) (document.UserRole, error)

	// IsUserAuthorized checks if a user is authorized to perform a specific action on a document
	// action could be "view", "edit", "delete", etc.
	IsUserAuthorized(ctx context.Context, docID document.DocumentID, userID string, action string) (bool, error)

	// AssignDocumentToSpace assigns a document to a space
	// Creates the many-to-many relationship
	AssignDocumentToSpace(ctx context.Context, docID document.DocumentID, spaceID string, assignedBy string) error

	// RemoveDocumentFromSpace removes a document from a space
	// Breaks the many-to-many relationship
	RemoveDocumentFromSpace(ctx context.Context, docID document.DocumentID, spaceID string) error

	// GetDocumentSpaces retrieves all spaces a document is assigned to
	GetDocumentSpaces(ctx context.Context, docID document.DocumentID) ([]string, error)

	// GetSpaceDocuments retrieves all documents assigned to a space
	GetSpaceDocuments(ctx context.Context, spaceID string, userID string, offset, limit int) ([]*document.DocumentMetadata, int, error)

	// GetUserRoleInSpace retrieves the user's role within a specific space.
	// This is crucial for space-level permission checks.
	GetUserRoleInSpace(ctx context.Context, userID string, spaceID string) (document.UserRole, error)

	// WithTransaction starts a transaction and returns a repository that operates within it
	// Used for operations that need to be atomic
	WithTransaction(ctx context.Context) (DocumentRepositoryPort, error)

	// Commit commits a transaction
	Commit(ctx context.Context) error

	// Rollback rolls back a transaction
	Rollback(ctx context.Context) error
}
