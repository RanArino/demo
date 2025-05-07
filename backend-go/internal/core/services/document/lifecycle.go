package document

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/ran/demo/backend-go/internal/core/models/document"
	ports "github.com/ran/demo/backend-go/internal/core/ports/document"
)

// DocumentLifecycleManager implements the DocumentLifecycleManagerPort interface
type DocumentLifecycleManager struct {
	repository ports.DocumentRepositoryPort
	storage    ports.StoragePort
	queue      ports.QueuePort
	parser     ports.FileParsingOrchestratorPort
	presenter  ports.DocumentPresenterPort
	logger     *slog.Logger
}

// NewDocumentLifecycleManager creates a new document lifecycle manager
func NewDocumentLifecycleManager(
	repository ports.DocumentRepositoryPort,
	storage ports.StoragePort,
	queue ports.QueuePort,
	parser ports.FileParsingOrchestratorPort,
	presenter ports.DocumentPresenterPort,
	logger *slog.Logger,
) *DocumentLifecycleManager {
	return &DocumentLifecycleManager{
		repository: repository,
		storage:    storage,
		queue:      queue,
		parser:     parser,
		presenter:  presenter,
		logger:     logger,
	}
}

// HandleFileUpload processes a direct file upload
func (m *DocumentLifecycleManager) HandleFileUpload(
	ctx context.Context,
	userID string,
	spaceID string,
	file *document.DocumentContent,
) (*document.DocumentMetadata, error) {
	// First, check if the user has permission to upload to this space
	authorized, err := m.isUserAuthorizedForSpace(ctx, userID, spaceID, "upload")
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, document.ErrStorageAccessDenied
	}

	// Create new document ID
	docID := document.DocumentID(uuid.New().String())

	// Create document metadata
	metadata := &document.DocumentMetadata{
		ID:               docID,
		OwnerID:          userID,
		OriginalFilename: file.OriginalFilename,
		Source:           "upload",
		MIMEType:         file.MIMEType,
		Size:             file.Size,
		Status:           document.StatusUploading,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Start a transaction for creating metadata and handling storage
	txRepo, err := m.repository.WithTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Store metadata first
	_, err = txRepo.CreateDocumentMetadata(ctx, metadata, userID)
	if err != nil {
		txRepo.Rollback(ctx)
		return nil, fmt.Errorf("failed to create document metadata: %w", err)
	}

	// Assign document to space
	err = txRepo.AssignDocumentToSpace(ctx, docID, spaceID, userID)
	if err != nil {
		txRepo.Rollback(ctx)
		return nil, fmt.Errorf("failed to assign document to space: %w", err)
	}

	// Upload file to storage
	storageOpts := &document.StorageOptions{
		ContentType: file.MIMEType,
		MetadataAttributes: map[string]string{
			"document_id": string(docID),
			"space_id":    spaceID,
			"owner_id":    userID,
			"filename":    file.OriginalFilename,
		},
	}

	storageObj, err := m.storage.Upload(ctx, file, storageOpts)
	if err != nil {
		txRepo.Rollback(ctx)
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	// Update metadata with storage reference
	metadata.OriginalFileReference = storageObj.Key
	metadata.Status = document.StatusPendingProcessing
	metadata.UpdatedAt = time.Now()

	err = txRepo.UpdateDocumentStatus(ctx, docID, document.StatusPendingProcessing)
	if err != nil {
		txRepo.Rollback(ctx)
		return nil, fmt.Errorf("failed to update document status: %w", err)
	}

	// Commit the transaction
	err = txRepo.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Queue document for processing
	parseTask := &document.DocumentParseTask{
		DocumentID:       docID,
		StorageObjectID:  storageObj.ID,
		UserID:           userID,
		UserTier:         "standard", // This would come from user service in a real implementation
		UserDomain:       "general",
		OriginalFilename: file.OriginalFilename,
		MIMEType:         file.MIMEType,
		Size:             file.Size,
	}

	err = m.queue.SendDocumentParseTask(ctx, parseTask)
	if err != nil {
		m.logger.Error("Failed to queue document for processing",
			"error", err,
			"document_id", docID,
			"user_id", userID)
		// Update status to error, but don't fail the request
		// The document was created and stored, but processing will need to be triggered manually
		_ = m.repository.UpdateDocumentStatus(ctx, docID, document.StatusError)
		metadata.Status = document.StatusError
		metadata.ErrorDetails = "Failed to queue document for processing"
	}

	return metadata, nil
}

// HandleExternalImport processes a document import from an external source
func (m *DocumentLifecycleManager) HandleExternalImport(
	ctx context.Context,
	userID string,
	importRequest *ports.ImportRequest,
) (*document.DocumentMetadata, error) {
	// Check permissions
	authorized, err := m.isUserAuthorizedForSpace(ctx, userID, importRequest.SpaceID, "upload")
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, document.ErrStorageAccessDenied
	}

	// For the initial implementation, external imports would require additional services
	// For simplicity, we'll just return an error indicating it's not implemented yet
	return nil, fmt.Errorf("external imports not implemented yet")
}

// GetDocumentDetails retrieves detailed metadata about a document
func (m *DocumentLifecycleManager) GetDocumentDetails(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
) (*document.DocumentMetadata, error) {
	// Fetch metadata, this also checks permissions
	metadata, err := m.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// GetDocumentContent retrieves the processed content of a document
func (m *DocumentLifecycleManager) GetDocumentContent(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
) (*document.StructuredDocumentOutput, error) {
	// Check if the document is ready
	metadata, err := m.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	if metadata.Status != document.StatusReady {
		return nil, fmt.Errorf("document is not ready: %s", metadata.Status)
	}

	// Fetch the processed content
	content, err := m.repository.GetProcessedContent(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// GetMarkdownContent retrieves just the markdown content of a document
func (m *DocumentLifecycleManager) GetMarkdownContent(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
) (string, error) {
	// This will check permissions
	markdown, err := m.repository.GetProcessedMarkdown(ctx, docID, userID)
	if err != nil {
		return "", err
	}

	return markdown, nil
}

// StreamOriginalContent streams the original document file to the provided writer
func (m *DocumentLifecycleManager) StreamOriginalContent(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	writer io.Writer,
) error {
	// Check permissions and get metadata
	metadata, err := m.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return err
	}

	// Check if the user can download
	authorized, err := m.repository.IsUserAuthorized(ctx, docID, userID, "download")
	if err != nil {
		return fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return document.ErrStorageAccessDenied
	}

	// Get storage object ID from the reference
	storageObjID := document.StorageObjectID(metadata.OriginalFileReference)

	// Stream content directly to the writer
	err = m.storage.DownloadToWriter(ctx, storageObjID, writer)
	if err != nil {
		return fmt.Errorf("failed to stream content: %w", err)
	}

	return nil
}

// GetDownloadURL generates a pre-signed URL for downloading the original document
func (m *DocumentLifecycleManager) GetDownloadURL(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
) (*document.StorageDownloadInfo, error) {
	// Check permissions and get metadata
	metadata, err := m.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	// Check if the user can download
	authorized, err := m.repository.IsUserAuthorized(ctx, docID, userID, "download")
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, document.ErrStorageAccessDenied
	}

	// Get storage object ID from the reference
	storageObjID := document.StorageObjectID(metadata.OriginalFileReference)

	// Generate pre-signed URL - expires in 15 minutes
	downloadInfo, err := m.storage.GetPresignedDownloadURL(ctx, storageObjID, metadata.OriginalFilename, 900)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return downloadInfo, nil
}

// GetPresignedUploadURL generates a pre-signed URL for direct upload
func (m *DocumentLifecycleManager) GetPresignedUploadURL(
	ctx context.Context,
	userID string,
	spaceID string,
	filename string,
	contentType string,
) (*document.StorageUploadInfo, error) {
	// Check permissions
	authorized, err := m.isUserAuthorizedForSpace(ctx, userID, spaceID, "upload")
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, document.ErrStorageAccessDenied
	}

	// Generate a document ID for this future upload
	docID := document.DocumentID(uuid.New().String())

	// Create storage options
	storageOpts := &document.StorageOptions{
		ContentType: contentType,
		MetadataAttributes: map[string]string{
			"document_id": string(docID),
			"space_id":    spaceID,
			"owner_id":    userID,
			"filename":    filename,
		},
	}

	// Generate pre-signed URL - expires in 30 minutes
	uploadInfo, err := m.storage.GetPresignedUploadURL(ctx, filename, contentType, storageOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	// Create a placeholder document record
	metadata := &document.DocumentMetadata{
		ID:               docID,
		OwnerID:          userID,
		OriginalFilename: filename,
		Source:           "presigned_upload",
		MIMEType:         contentType,
		Size:             0, // Will be updated after upload
		Status:           document.StatusUploading,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Start a transaction for creating metadata and space assignment
	txRepo, err := m.repository.WithTransaction(ctx)
	if err != nil {
		// Log but don't fail - the client will still be able to upload
		m.logger.Error("Failed to start transaction for document creation",
			"error", err,
			"document_id", docID,
			"user_id", userID)
		return uploadInfo, nil
	}

	// Store metadata
	_, err = txRepo.CreateDocumentMetadata(ctx, metadata, userID)
	if err != nil {
		// Log but don't fail - the client will still be able to upload
		txRepo.Rollback(ctx)
		m.logger.Error("Failed to create placeholder document metadata",
			"error", err,
			"document_id", docID,
			"user_id", userID)
		return uploadInfo, nil
	}

	// Assign document to space
	err = txRepo.AssignDocumentToSpace(ctx, docID, spaceID, userID)
	if err != nil {
		// Log but don't fail - the client will still be able to upload
		txRepo.Rollback(ctx)
		m.logger.Error("Failed to assign document to space",
			"error", err,
			"document_id", docID,
			"space_id", spaceID,
			"user_id", userID)
		return uploadInfo, nil
	}

	// Commit the transaction
	err = txRepo.Commit(ctx)
	if err != nil {
		// Log but don't fail - the client will still be able to upload
		m.logger.Error("Failed to commit transaction for document creation",
			"error", err,
			"document_id", docID,
			"user_id", userID)
	}

	return uploadInfo, nil
}

// DeleteDocument deletes a document
func (m *DocumentLifecycleManager) DeleteDocument(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	permanent bool,
) error {
	// Check permissions - different permissions for permanent vs soft delete
	action := "delete"
	if permanent {
		action = "permanent_delete"
	}

	authorized, err := m.repository.IsUserAuthorized(ctx, docID, userID, action)
	if err != nil {
		return fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return document.ErrStorageAccessDenied
	}

	if permanent {
		// Get metadata first to get the storage references
		metadata, err := m.repository.GetDocumentMetadata(ctx, docID, userID)
		if err != nil {
			return err
		}

		// Delete the original file from storage
		if metadata.OriginalFileReference != "" {
			storageObjID := document.StorageObjectID(metadata.OriginalFileReference)
			err = m.storage.Delete(ctx, storageObjID)
			if err != nil {
				m.logger.Error("Failed to delete original file from storage",
					"error", err,
					"document_id", docID,
					"storage_reference", metadata.OriginalFileReference)
				// Continue with deletion even if storage delete fails
			}
		}

		// Delete the processed content if it exists
		if metadata.ProcessedContentReference != "" {
			storageObjID := document.StorageObjectID(metadata.ProcessedContentReference)
			err = m.storage.Delete(ctx, storageObjID)
			if err != nil {
				m.logger.Error("Failed to delete processed content from storage",
					"error", err,
					"document_id", docID,
					"storage_reference", metadata.ProcessedContentReference)
				// Continue with deletion even if storage delete fails
			}
		}

		// Permanently delete from repository
		err = m.repository.PermanentlyDeleteDocument(ctx, docID, userID)
		if err != nil {
			return fmt.Errorf("failed to permanently delete document: %w", err)
		}
	} else {
		// Soft delete
		err = m.repository.DeleteDocument(ctx, docID, userID)
		if err != nil {
			return fmt.Errorf("failed to delete document: %w", err)
		}
	}

	return nil
}

// ReprocessDocument forces reprocessing of a document
func (m *DocumentLifecycleManager) ReprocessDocument(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	options *ports.FileParsingOptions,
) (*document.DocumentMetadata, error) {
	// Check permissions
	authorized, err := m.repository.IsUserAuthorized(ctx, docID, userID, "process")
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, document.ErrStorageAccessDenied
	}

	// Get document metadata
	metadata, err := m.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	// Only allow reprocessing if the document is in a terminal state
	if metadata.Status == document.StatusUploading ||
		metadata.Status == document.StatusPendingProcessing ||
		metadata.Status == document.StatusParsing ||
		metadata.Status == document.StatusProcessing {
		return nil, document.ErrDocumentProcessing
	}

	// Update status to pending processing
	err = m.repository.UpdateDocumentStatus(ctx, docID, document.StatusPendingProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to update document status: %w", err)
	}

	// Get storage object ID from the reference
	storageObjID := document.StorageObjectID(metadata.OriginalFileReference)

	// Queue document for processing
	parseTask := &document.DocumentParseTask{
		DocumentID:       docID,
		StorageObjectID:  storageObjID,
		UserID:           userID,
		UserTier:         "standard", // This would come from user service in a real implementation
		UserDomain:       "general",
		OriginalFilename: metadata.OriginalFilename,
		MIMEType:         metadata.MIMEType,
		Size:             metadata.Size,
	}

	err = m.queue.SendDocumentParseTask(ctx, parseTask)
	if err != nil {
		// Revert status
		_ = m.repository.UpdateDocumentStatus(ctx, docID, document.StatusError)
		return nil, fmt.Errorf("failed to queue document for processing: %w", err)
	}

	// Return updated metadata
	metadata.Status = document.StatusPendingProcessing
	metadata.UpdatedAt = time.Now()
	metadata.ProcessingAttempts++

	return metadata, nil
}

// ListUserDocuments lists documents the user has access to
func (m *DocumentLifecycleManager) ListUserDocuments(
	ctx context.Context,
	userID string,
	spaceID *string,
	offset, limit int,
) ([]*document.DocumentMetadata, int, error) {
	// Repository handles permission filtering
	return m.repository.ListUserDocuments(ctx, userID, spaceID, offset, limit)
}

// CanUserPerformAction checks if a user can perform an action on a document
func (m *DocumentLifecycleManager) CanUserPerformAction(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	action string,
) (bool, error) {
	return m.repository.IsUserAuthorized(ctx, docID, userID, action)
}

// isUserAuthorizedForSpace checks if a user is authorized to perform an action in a space
// This is a helper method for space-level operations (not document-specific)
func (m *DocumentLifecycleManager) isUserAuthorizedForSpace(
	ctx context.Context,
	userID string,
	spaceID string,
	action string,
) (bool, error) {
	// In a real implementation, this would call a space service or use a space repository
	// For the demo, we'll just assume the user is an Owner (for simplicity)
	// This allows us to test the upload flow without implementing space permissions

	// Normally we'd do something like:
	// role, err := spaceService.GetUserRole(ctx, userID, spaceID)
	// if err != nil {
	//     return false, err
	// }

	// For demo purposes, assume everyone is an Owner
	role := document.RoleOwner

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
