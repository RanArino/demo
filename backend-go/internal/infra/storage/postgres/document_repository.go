package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ran/demo/backend-go/internal/core/models/document"
	ports "github.com/ran/demo/backend-go/internal/core/ports/document"
)

// DocumentRepository implements document.DocumentRepositoryPort for PostgreSQL
type DocumentRepository struct {
	db     *sqlx.DB
	tx     *sqlx.Tx
	logger *slog.Logger
}

// Helper methods for ExtContext
func (r *DocumentRepository) getContext() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *DocumentRepository) get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	db := r.getContext()
	if tx, ok := db.(*sqlx.Tx); ok {
		return tx.GetContext(ctx, dest, query, args...)
	}
	if db, ok := db.(*sqlx.DB); ok {
		return db.GetContext(ctx, dest, query, args...)
	}
	return fmt.Errorf("unknown context type")
}

func (r *DocumentRepository) select_(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	db := r.getContext()
	if tx, ok := db.(*sqlx.Tx); ok {
		return tx.SelectContext(ctx, dest, query, args...)
	}
	if db, ok := db.(*sqlx.DB); ok {
		return db.SelectContext(ctx, dest, query, args...)
	}
	return fmt.Errorf("unknown context type")
}

// NewDocumentRepository creates a new PostgreSQL document repository
func NewDocumentRepository(db *sqlx.DB, logger *slog.Logger) *DocumentRepository {
	return &DocumentRepository{
		db:     db,
		logger: logger,
	}
}

// CreateDocumentMetadata creates a new document metadata entry
func (r *DocumentRepository) CreateDocumentMetadata(
	ctx context.Context,
	metadata *document.DocumentMetadata,
	userID string,
) (document.DocumentID, error) {
	query := `
		INSERT INTO documents (
			id, owner_id, original_filename, source, mime_type, 
			size, status, original_file_reference, processed_content_reference,
			error_details, processing_attempts, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id`

	db := r.getContext()
	var id string
	err := db.QueryRowxContext(
		ctx,
		query,
		metadata.ID,
		metadata.OwnerID,
		metadata.OriginalFilename,
		metadata.Source,
		metadata.MIMEType,
		metadata.Size,
		metadata.Status,
		metadata.OriginalFileReference,
		metadata.ProcessedContentReference,
		metadata.ErrorDetails,
		metadata.ProcessingAttempts,
		metadata.CreatedAt,
		metadata.UpdatedAt,
	).Scan(&id)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" { // Unique violation
				return "", document.ErrDocumentAlreadyExists
			}
		}
		return "", fmt.Errorf("failed to create document metadata: %w", err)
	}

	// Set up initial document permissions
	err = r.createInitialPermissions(ctx, string(metadata.ID), userID)
	if err != nil {
		return "", fmt.Errorf("failed to set up document permissions: %w", err)
	}

	return document.DocumentID(id), nil
}

// createInitialPermissions sets up initial document permissions for the creator
func (r *DocumentRepository) createInitialPermissions(
	ctx context.Context,
	docID string,
	creatorID string,
) error {
	// In a complete implementation, we'd query the space membership table
	// to determine who has access to the document based on space roles.
	// For now, we'll just set up the creator as an owner.

	db := r.getContext()
	query := `
		INSERT INTO document_permissions (
			document_id, user_id, role, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $4
		)`

	_, err := db.ExecContext(
		ctx,
		query,
		docID,
		creatorID,
		document.RoleOwner,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create initial permission: %w", err)
	}

	return nil
}

// GetDocumentMetadata retrieves document metadata by ID
func (r *DocumentRepository) GetDocumentMetadata(
	ctx context.Context,
	docID document.DocumentID,
	userID string,
) (*document.DocumentMetadata, error) {
	query := `
		SELECT d.* FROM documents d
		JOIN document_permissions p ON d.id = p.document_id
		WHERE d.id = $1 AND p.user_id = $2 AND d.deleted_at IS NULL`

	var metadata document.DocumentMetadata
	err := r.get(ctx, &metadata, query, docID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, document.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("failed to get document metadata: %w", err)
	}

	return &metadata, nil
}

// UpdateDocumentStatus updates the processing status of a document
func (r *DocumentRepository) UpdateDocumentStatus(
	ctx context.Context,
	docID document.DocumentID,
	status document.DocumentStatus,
) error {
	query := `
		UPDATE documents
		SET status = $1, updated_at = $2
		WHERE id = $3`

	db := r.getContext()
	_, err := db.ExecContext(ctx, query, status, time.Now(), docID)
	if err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	return nil
}

// StoreProcessedContent saves the structured document output after processing
func (r *DocumentRepository) StoreProcessedContent(
	ctx context.Context,
	docID document.DocumentID,
	structuredOutput *document.StructuredDocumentOutput,
) error {
	// In a complete implementation, we might store the content in JSON format in a document_content table
	// or serialize to a storage service. For simplicity, we'll update the documents table.
	query := `
		UPDATE documents
		SET 
			processed_content_reference = $1,
			processing_completed_at = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5`

	contentRef := ""
	if structuredOutput.StorageReference != "" {
		contentRef = structuredOutput.StorageReference
	}

	db := r.getContext()
	now := time.Now()
	_, err := db.ExecContext(
		ctx,
		query,
		contentRef,
		now,
		document.StatusReady,
		now,
		docID,
	)
	if err != nil {
		return fmt.Errorf("failed to store processed content: %w", err)
	}

	// Store document content in document_content table
	contentQuery := `
		INSERT INTO document_content (
			document_id, markdown_content, tables_json, images_json, metadata_json,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $6
		) ON CONFLICT (document_id) 
		DO UPDATE SET
			markdown_content = $2,
			tables_json = $3,
			images_json = $4,
			metadata_json = $5,
			updated_at = $6
	`

	// Marshal tables, images, and metadata to JSON if provided
	var tablesJSON, imagesJSON, metadataJSON []byte
	var err1, err2, err3 error

	if len(structuredOutput.Tables) > 0 {
		tablesJSON, err1 = json.Marshal(structuredOutput.Tables)
		if err1 != nil {
			return fmt.Errorf("failed to marshal tables to JSON: %w", err1)
		}
	}

	if len(structuredOutput.Images) > 0 {
		imagesJSON, err2 = json.Marshal(structuredOutput.Images)
		if err2 != nil {
			return fmt.Errorf("failed to marshal images to JSON: %w", err2)
		}
	}

	if structuredOutput.OtherMetadata != nil {
		metadataJSON, err3 = json.Marshal(structuredOutput.OtherMetadata)
		if err3 != nil {
			return fmt.Errorf("failed to marshal metadata to JSON: %w", err3)
		}
	}

	_, err = db.ExecContext(
		ctx,
		contentQuery,
		docID,
		structuredOutput.MarkdownContent,
		string(tablesJSON),
		string(imagesJSON),
		string(metadataJSON),
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to store document content: %w", err)
	}

	return nil
}

// GetProcessedContent retrieves the full structured document output
func (r *DocumentRepository) GetProcessedContent(
	ctx context.Context,
	docID document.DocumentID,
	userID string,
) (*document.StructuredDocumentOutput, error) {
	// Check if user has access to the document
	_, err := r.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	// Get processed content from document_content table
	query := `
		SELECT 
			markdown_content, tables_json, images_json, metadata_json
		FROM document_content
		WHERE document_id = $1`

	var output document.StructuredDocumentOutput
	err = r.get(ctx, &output, query, docID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, document.ErrContentNotProcessed
		}
		return nil, fmt.Errorf("failed to get processed content: %w", err)
	}

	// Get the storage reference
	refQuery := `
		SELECT processed_content_reference
		FROM documents
		WHERE id = $1`

	var storageRef string
	err = r.get(ctx, &storageRef, refQuery, docID)
	if err != nil {
		r.logger.Warn("Failed to get storage reference", "error", err, "document_id", docID)
		// Continue without the storage reference
	} else {
		output.StorageReference = storageRef
	}

	return &output, nil
}

// GetProcessedMarkdown retrieves just the markdown content of a processed document
func (r *DocumentRepository) GetProcessedMarkdown(
	ctx context.Context,
	docID document.DocumentID,
	userID string,
) (string, error) {
	// Check if user has access to the document
	_, err := r.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return "", err
	}

	// Get processed markdown content
	query := `
		SELECT markdown_content
		FROM document_content
		WHERE document_id = $1`

	var markdownContent string
	err = r.get(ctx, &markdownContent, query, docID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", document.ErrContentNotProcessed
		}
		return "", fmt.Errorf("failed to get markdown content: %w", err)
	}

	return markdownContent, nil
}

// ListUserDocuments lists documents the user has access to with pagination
func (r *DocumentRepository) ListUserDocuments(
	ctx context.Context,
	userID string,
	spaceID *string,
	offset, limit int,
) ([]*document.DocumentMetadata, int, error) {
	baseQuery := `
		FROM documents d
		JOIN document_permissions p ON d.id = p.document_id
		WHERE p.user_id = $1 AND d.deleted_at IS NULL`

	params := []interface{}{userID}
	paramIndex := 2

	// Add space filter if provided
	if spaceID != nil && *spaceID != "" {
		baseQuery += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1 FROM document_space_assignments dsa 
				WHERE dsa.document_id = d.id AND dsa.space_id = $%d
			)`, paramIndex)
		params = append(params, *spaceID)
		paramIndex++
	}

	countQuery := `SELECT COUNT(*) ` + baseQuery
	selectQuery := `SELECT d.* ` + baseQuery

	// Add pagination
	selectQuery = selectQuery + fmt.Sprintf(" ORDER BY d.created_at DESC LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
	params = append(params, limit, offset)

	// Get total count
	var totalCount int
	err := r.get(ctx, &totalCount, countQuery, params[:paramIndex-1]...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get documents
	var documents []*document.DocumentMetadata
	err = r.select_(ctx, &documents, selectQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}

	return documents, totalCount, nil
}

// DeleteDocument marks a document as deleted (soft delete)
func (r *DocumentRepository) DeleteDocument(
	ctx context.Context,
	docID document.DocumentID,
	userID string,
) error {
	// Check if user has permission to delete
	authorized, err := r.IsUserAuthorized(ctx, docID, userID, "delete")
	if err != nil {
		return err
	}
	if !authorized {
		return document.ErrPermissionDenied
	}

	// Soft delete the document
	query := `
		UPDATE documents
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2`

	db := r.getContext()
	_, err = db.ExecContext(ctx, query, time.Now(), docID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// PermanentlyDeleteDocument physically removes the document and its data
func (r *DocumentRepository) PermanentlyDeleteDocument(
	ctx context.Context,
	docID document.DocumentID,
	userID string,
) error {
	// Check if user has permission to permanently delete
	authorized, err := r.IsUserAuthorized(ctx, docID, userID, "permanent_delete")
	if err != nil {
		return err
	}
	if !authorized {
		return document.ErrPermissionDenied
	}

	// Start transaction if not already in one
	var tx *sqlx.Tx
	if r.tx != nil {
		tx = r.tx
	} else {
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()
	}

	// Delete document content
	contentQuery := `DELETE FROM document_content WHERE document_id = $1`
	_, err = tx.ExecContext(ctx, contentQuery, docID)
	if err != nil {
		return fmt.Errorf("failed to delete document content: %w", err)
	}

	// Delete document permissions
	permQuery := `DELETE FROM document_permissions WHERE document_id = $1`
	_, err = tx.ExecContext(ctx, permQuery, docID)
	if err != nil {
		return fmt.Errorf("failed to delete document permissions: %w", err)
	}

	// Delete document space assignments
	assignQuery := `DELETE FROM document_space_assignments WHERE document_id = $1`
	_, err = tx.ExecContext(ctx, assignQuery, docID)
	if err != nil {
		return fmt.Errorf("failed to delete document space assignments: %w", err)
	}

	// Delete document
	docQuery := `DELETE FROM documents WHERE id = $1`
	_, err = tx.ExecContext(ctx, docQuery, docID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Commit transaction if we started it
	if r.tx == nil {
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// GetUserRoleForDocument retrieves the user's role for a specific document
func (r *DocumentRepository) GetUserRoleForDocument(
	ctx context.Context,
	docID document.DocumentID,
	userID string,
) (document.UserRole, error) {
	query := `
		SELECT role
		FROM document_permissions
		WHERE document_id = $1 AND user_id = $2`

	var roleStr string
	err := r.get(ctx, &roleStr, query, docID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no explicit role, check space-level role
			return r.getSpaceRoleForDocument(ctx, docID, userID)
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	return document.UserRole(roleStr), nil
}

// getSpaceRoleForDocument gets the user's role in the document's space
func (r *DocumentRepository) getSpaceRoleForDocument(
	ctx context.Context,
	docID document.DocumentID,
	userID string,
) (document.UserRole, error) {
	query := `
		SELECT sm.role
		FROM document_space_assignments dsa
		JOIN space_members sm ON dsa.space_id = sm.space_id
		WHERE dsa.document_id = $1 AND sm.user_id = $2
		LIMIT 1`

	var roleStr string
	err := r.get(ctx, &roleStr, query, docID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return document.RoleGuest, nil // Default role if no explicit role
		}
		return "", fmt.Errorf("failed to get space role: %w", err)
	}

	return document.UserRole(roleStr), nil
}

// IsUserAuthorized checks if a user is authorized to perform a specific action on a document
func (r *DocumentRepository) IsUserAuthorized(
	ctx context.Context,
	docID document.DocumentID,
	userID string,
	action string,
) (bool, error) {
	// Get user's role for this document
	role, err := r.GetUserRoleForDocument(ctx, docID, userID)
	if err != nil {
		return false, err
	}

	// Check if the role has the required permission
	permissions, exists := document.RolePermissions[role]
	if !exists {
		return false, fmt.Errorf("invalid role: %s", role)
	}

	allowed, exists := permissions[action]
	if !exists {
		return false, fmt.Errorf("unknown action: %s", action)
	}

	return allowed, nil
}

// WithTransaction starts a transaction and returns a repository that operates within it
func (r *DocumentRepository) WithTransaction(ctx context.Context) (ports.DocumentRepositoryPort, error) {
	if r.tx != nil {
		return r, nil // Already in a transaction
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	return &DocumentRepository{
		db:     r.db,
		tx:     tx,
		logger: r.logger,
	}, nil
}

// Commit commits a transaction
func (r *DocumentRepository) Commit(ctx context.Context) error {
	if r.tx == nil {
		return fmt.Errorf("no active transaction to commit")
	}

	err := r.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back a transaction
func (r *DocumentRepository) Rollback(ctx context.Context) error {
	if r.tx == nil {
		return fmt.Errorf("no active transaction to rollback")
	}

	err := r.tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

// AssignDocumentToSpace assigns a document to a space
func (r *DocumentRepository) AssignDocumentToSpace(
	ctx context.Context,
	docID document.DocumentID,
	spaceID string,
	assignedBy string,
) error {
	query := `
		INSERT INTO document_space_assignments (
			document_id, space_id, assigned_at, assigned_by
		) VALUES (
			$1, $2, $3, $4
		) ON CONFLICT (document_id, space_id) DO NOTHING`

	db := r.getContext()
	_, err := db.ExecContext(
		ctx,
		query,
		docID,
		spaceID,
		time.Now(),
		assignedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to assign document to space: %w", err)
	}

	return nil
}

// RemoveDocumentFromSpace removes a document from a space
func (r *DocumentRepository) RemoveDocumentFromSpace(
	ctx context.Context,
	docID document.DocumentID,
	spaceID string,
) error {
	query := `
		DELETE FROM document_space_assignments
		WHERE document_id = $1 AND space_id = $2`

	db := r.getContext()
	_, err := db.ExecContext(
		ctx,
		query,
		docID,
		spaceID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove document from space: %w", err)
	}

	return nil
}

// GetDocumentSpaces retrieves all spaces a document is assigned to
func (r *DocumentRepository) GetDocumentSpaces(
	ctx context.Context,
	docID document.DocumentID,
) ([]string, error) {
	query := `
		SELECT space_id
		FROM document_space_assignments
		WHERE document_id = $1`

	var spaceIDs []string
	err := r.select_(ctx, &spaceIDs, query, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document spaces: %w", err)
	}

	return spaceIDs, nil
}

// GetSpaceDocuments retrieves all documents assigned to a space
func (r *DocumentRepository) GetSpaceDocuments(
	ctx context.Context,
	spaceID string,
	userID string,
	offset, limit int,
) ([]*document.DocumentMetadata, int, error) {
	baseQuery := `
		FROM documents d
		JOIN document_space_assignments dsa ON d.id = dsa.document_id
		JOIN document_permissions p ON d.id = p.document_id
		WHERE dsa.space_id = $1 AND p.user_id = $2 AND d.deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) ` + baseQuery
	selectQuery := `SELECT d.* ` + baseQuery + ` ORDER BY d.created_at DESC LIMIT $3 OFFSET $4`

	// Get total count
	var totalCount int
	err := r.get(ctx, &totalCount, countQuery, spaceID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count space documents: %w", err)
	}

	// Get documents
	var documents []*document.DocumentMetadata
	err = r.select_(ctx, &documents, selectQuery, spaceID, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list space documents: %w", err)
	}

	return documents, totalCount, nil
}

// GetUserRoleInSpace retrieves the user's role within a specific space.
func (r *DocumentRepository) GetUserRoleInSpace(
	ctx context.Context,
	userID string,
	spaceID string,
) (document.UserRole, error) {
	query := `
		SELECT role
		FROM space_members
		WHERE space_id = $1 AND user_id = $2 AND status = 'active'`

	var roleStr string
	err := r.get(ctx, &roleStr, query, spaceID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If user is the owner, they should have owner role even if not explicitly in space_members
			ownerQuery := `
				SELECT EXISTS (
					SELECT 1 FROM spaces
					WHERE id = $1 AND owner_id = $2
				)`

			var isOwner bool
			err := r.get(ctx, &isOwner, ownerQuery, spaceID, userID)
			if err != nil {
				r.logger.Error("Error checking if user is space owner", "error", err, "space_id", spaceID, "user_id", userID)
				return document.RoleGuest, nil
			}

			if isOwner {
				return document.RoleOwner, nil
			}

			// Check for guest access
			guestQuery := `
				SELECT guest_access_enabled
				FROM spaces
				WHERE id = $1 AND guest_access_enabled = true 
				AND (guest_access_expiry IS NULL OR guest_access_expiry > NOW())`

			var guestEnabled bool
			err = r.get(ctx, &guestEnabled, guestQuery, spaceID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				r.logger.Error("Error checking guest access", "error", err, "space_id", spaceID)
			}

			if guestEnabled {
				return document.RoleGuest, nil
			}

			// If no explicit role in the space, default to Guest or a more restrictive role.
			// Alternatively, could return an error if user must be an explicit member.
			return document.RoleGuest, nil
		}
		return "", fmt.Errorf("failed to get user role in space: %w", err)
	}

	return document.UserRole(roleStr), nil
}
