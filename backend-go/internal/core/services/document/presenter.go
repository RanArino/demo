package document

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/ran/demo/backend-go/internal/core/models/document"
	ports "github.com/ran/demo/backend-go/internal/core/ports/document"
)

// DocumentPresenter implements the DocumentPresenterPort interface
type DocumentPresenter struct {
	repository ports.DocumentRepositoryPort
	storage    ports.StoragePort
	logger     *slog.Logger
}

// NewDocumentPresenter creates a new document presenter
func NewDocumentPresenter(
	repository ports.DocumentRepositoryPort,
	storage ports.StoragePort,
	logger *slog.Logger,
) *DocumentPresenter {
	return &DocumentPresenter{
		repository: repository,
		storage:    storage,
		logger:     logger,
	}
}

// GetMarkdownPreview generates a preview of the document in markdown format
func (p *DocumentPresenter) GetMarkdownPreview(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	options *ports.PreviewOptions,
) (string, error) {
	// Check user access
	metadata, err := p.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return "", err
	}

	// Check if document is ready
	if metadata.Status != document.StatusReady {
		return "", fmt.Errorf("document is not ready for preview: %s", metadata.Status)
	}

	// Get the markdown content
	content, err := p.repository.GetProcessedMarkdown(ctx, docID, userID)
	if err != nil {
		return "", err
	}

	// Apply preview options
	if options != nil {
		// Truncate to max length if specified
		if options.MaxPreviewLength > 0 && len(content) > options.MaxPreviewLength {
			content = content[:options.MaxPreviewLength] + "..."
		}

		// Highlight terms if specified
		if len(options.HighlightTerms) > 0 {
			for _, term := range options.HighlightTerms {
				content = strings.ReplaceAll(
					content,
					term,
					fmt.Sprintf("**%s**", term),
				)
			}
		}
	}

	return content, nil
}

// GetOriginalDownloadable prepares the original document for download
func (p *DocumentPresenter) GetOriginalDownloadable(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
) (*document.FileDownload, error) {
	// Check user access
	metadata, err := p.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	// Check if user can download
	authorized, err := p.repository.IsUserAuthorized(ctx, docID, userID, "download")
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, document.ErrPermissionDenied
	}

	// Get storage object ID from the reference
	storageObjID := document.StorageObjectID(metadata.OriginalFileReference)

	// Get the document content
	content, err := p.storage.Download(ctx, storageObjID)
	if err != nil {
		return nil, err
	}

	// Create file download
	fileDownload := &document.FileDownload{
		ContentType:     content.MIMEType,
		FileName:        metadata.OriginalFilename,
		ContentLength:   content.Size,
		Content:         content.Stream,
		ContentProvider: "s3", // Assumes S3 storage
	}

	return fileDownload, nil
}

// GetDownloadable prepares a document for download in a specified format
func (p *DocumentPresenter) GetDownloadable(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	format ports.DownloadFormat,
) (*document.FileDownload, error) {
	// Check user access
	metadata, err := p.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	// Check if user can download
	authorized, err := p.repository.IsUserAuthorized(ctx, docID, userID, "download")
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, document.ErrPermissionDenied
	}

	// Handle different formats
	switch format {
	case ports.DownloadFormatOriginal:
		return p.GetOriginalDownloadable(ctx, userID, docID)

	case ports.DownloadFormatMarkdown:
		// Get the processed content
		content, err := p.repository.GetProcessedMarkdown(ctx, docID, userID)
		if err != nil {
			return nil, err
		}

		// Create file download with markdown content
		fileDownload := &document.FileDownload{
			ContentType:     "text/markdown",
			FileName:        strings.TrimSuffix(metadata.OriginalFilename, ".") + ".md",
			ContentLength:   int64(len(content)),
			Content:         strings.NewReader(content),
			ContentProvider: "memory",
		}

		return fileDownload, nil

	case ports.DownloadFormatText:
		// Get the processed content
		structContent, err := p.repository.GetProcessedContent(ctx, docID, userID)
		if err != nil {
			return nil, err
		}

		// Use raw text if available, otherwise use markdown
		textContent := structContent.RawText
		if textContent == "" {
			textContent = structContent.MarkdownContent
		}

		// Create file download with text content
		fileDownload := &document.FileDownload{
			ContentType:     "text/plain",
			FileName:        strings.TrimSuffix(metadata.OriginalFilename, ".") + ".txt",
			ContentLength:   int64(len(textContent)),
			Content:         strings.NewReader(textContent),
			ContentProvider: "memory",
		}

		return fileDownload, nil

	default:
		return nil, fmt.Errorf("unsupported download format: %s", format)
	}
}

// StreamContent streams the document content to the provided writer
func (p *DocumentPresenter) StreamContent(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	format ports.DownloadFormat,
	writer io.Writer,
) error {
	// Get the downloadable
	downloadable, err := p.GetDownloadable(ctx, userID, docID, format)
	if err != nil {
		return err
	}

	// Stream the content to the writer
	_, err = io.Copy(writer, downloadable.Content)
	if err != nil {
		return fmt.Errorf("failed to stream content: %w", err)
	}

	return nil
}

// SetupOriginalPreview generates pre-signed URLs for original file previews
func (p *DocumentPresenter) SetupOriginalPreview(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
) (*document.StorageDownloadInfo, error) {
	// Check user access
	metadata, err := p.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return nil, err
	}

	// Check if user can view
	authorized, err := p.repository.IsUserAuthorized(ctx, docID, userID, "view")
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}
	if !authorized {
		return nil, document.ErrPermissionDenied
	}

	// Get storage object ID from the reference
	storageObjID := document.StorageObjectID(metadata.OriginalFileReference)

	// Generate pre-signed URL - expires in 15 minutes
	downloadInfo, err := p.storage.GetPresignedDownloadURL(ctx, storageObjID, metadata.OriginalFilename, 900)
	if err != nil {
		return nil, fmt.Errorf("failed to generate preview URL: %w", err)
	}

	return downloadInfo, nil
}

// GenerateTablePreview generates an HTML preview of a table extracted from a document
func (p *DocumentPresenter) GenerateTablePreview(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	tableIndex int,
) (string, error) {
	// Check user access
	_, err := p.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return "", err
	}

	// Get the processed content
	content, err := p.repository.GetProcessedContent(ctx, docID, userID)
	if err != nil {
		return "", err
	}

	// Check if the table exists
	if tableIndex < 0 || tableIndex >= len(content.Tables) {
		return "", fmt.Errorf("table index %d out of range", tableIndex)
	}

	// Generate HTML for the table
	table := content.Tables[tableIndex]
	html := "<table>\n"

	// Add headers if any
	if len(table.Headers) > 0 {
		html += "  <thead>\n    <tr>\n"
		for _, header := range table.Headers {
			html += fmt.Sprintf("      <th>%s</th>\n", header)
		}
		html += "    </tr>\n  </thead>\n"
	}

	// Add rows
	html += "  <tbody>\n"
	for _, row := range table.Rows {
		html += "    <tr>\n"
		for _, cell := range row {
			html += fmt.Sprintf("      <td>%s</td>\n", cell)
		}
		html += "    </tr>\n"
	}
	html += "  </tbody>\n</table>"

	return html, nil
}

// GenerateImagePreview generates a preview URL for an image extracted from a document
func (p *DocumentPresenter) GenerateImagePreview(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	imageIndex int,
) (string, error) {
	// In a full implementation, this would handle generating preview URLs for extracted images
	// For now, return an error indicating it's not implemented
	return "", fmt.Errorf("image preview generation not implemented yet")
}

// GetDocumentSummary retrieves or generates a summary of the document
func (p *DocumentPresenter) GetDocumentSummary(
	ctx context.Context,
	userID string,
	docID document.DocumentID,
	maxLength int,
) (string, error) {
	// Check user access
	metadata, err := p.repository.GetDocumentMetadata(ctx, docID, userID)
	if err != nil {
		return "", err
	}

	// If the document already has a summary, return it
	if metadata.Summary != "" {
		// Truncate if necessary
		if maxLength > 0 && len(metadata.Summary) > maxLength {
			return metadata.Summary[:maxLength] + "...", nil
		}
		return metadata.Summary, nil
	}

	// In a full implementation, this would generate a summary using ML
	// For now, just return a generic message
	return "Document summary not available. Use document preprocessing to generate a summary.", nil
}
