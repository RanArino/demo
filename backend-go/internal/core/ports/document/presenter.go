package document

import (
	"context"
	"io"

	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// DownloadFormat represents the format for document downloads
type DownloadFormat string

const (
	DownloadFormatOriginal DownloadFormat = "ORIGINAL" // Original file format
	DownloadFormatMarkdown DownloadFormat = "MARKDOWN" // Markdown format
	DownloadFormatPDF      DownloadFormat = "PDF"      // PDF format
	DownloadFormatHTML     DownloadFormat = "HTML"     // HTML format
	DownloadFormatText     DownloadFormat = "TEXT"     // Plain text format
)

// PreviewOptions contains configuration for document previews
type PreviewOptions struct {
	MaxPreviewLength int      // Maximum length of preview content
	PreviewFormat    string   // Format of the preview (e.g., "html", "text")
	HighlightTerms   []string // Terms to highlight in the preview
	IncludeTables    bool     // Whether to include tables in the preview
	IncludeImages    bool     // Whether to include images in the preview
}

// DocumentPresenterPort prepares documents for client consumption
type DocumentPresenterPort interface {
	// GetMarkdownPreview generates a preview of the document in markdown format
	// Verifies user has permission to view the document (Viewer+)
	GetMarkdownPreview(ctx context.Context, userID string, docID document.DocumentID, options *PreviewOptions) (string, error)

	// GetOriginalDownloadable prepares the original document for download
	// Verifies user has permission to view the document (Viewer+)
	GetOriginalDownloadable(ctx context.Context, userID string, docID document.DocumentID) (*document.FileDownload, error)

	// GetDownloadable prepares a document for download in a specified format
	// Verifies user has permission to view the document (Viewer+)
	GetDownloadable(ctx context.Context, userID string, docID document.DocumentID, format DownloadFormat) (*document.FileDownload, error)

	// StreamContent streams the document content to the provided writer
	// Verifies user has permission to view the document (Viewer+)
	StreamContent(ctx context.Context, userID string, docID document.DocumentID, format DownloadFormat, writer io.Writer) error

	// SetupOriginalPreview generates pre-signed URLs for original file previews
	// Verifies user has permission to view the document (Viewer+)
	SetupOriginalPreview(ctx context.Context, userID string, docID document.DocumentID) (*document.StorageDownloadInfo, error)

	// GenerateTablePreview generates an HTML preview of a table extracted from a document
	// Verifies user has permission to view the document (Viewer+)
	GenerateTablePreview(ctx context.Context, userID string, docID document.DocumentID, tableIndex int) (string, error)

	// GenerateImagePreview generates a preview URL for an image extracted from a document
	// Verifies user has permission to view the document (Viewer+)
	GenerateImagePreview(ctx context.Context, userID string, docID document.DocumentID, imageIndex int) (string, error)

	// GetDocumentSummary retrieves or generates a summary of the document
	// Verifies user has permission to view the document (Viewer+)
	GetDocumentSummary(ctx context.Context, userID string, docID document.DocumentID, maxLength int) (string, error)
}
