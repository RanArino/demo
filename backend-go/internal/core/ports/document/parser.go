package document

import (
	"context"

	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// FileParsingOptions contains configuration for document parsing
type FileParsingOptions struct {
	ExtractTables     bool                   // Whether to extract tables from the document
	ExtractForms      bool                   // Whether to extract forms/structured data
	ExtractImages     bool                   // Whether to extract and store images
	PreferredStrategy string                 // Optional preferred parsing strategy
	Language          string                 // Optional document language hint
	ConversionOptions map[string]interface{} // Format conversion options
	StrategyOptions   map[string]interface{} // Strategy-specific options
}

// FileParsingOrchestratorPort orchestrates document parsing using appropriate strategies
type FileParsingOrchestratorPort interface {
	// ParseDocument parses a document using the appropriate strategy
	// The orchestrator will select the best strategy based on file type and user context
	ParseDocument(ctx context.Context, content *document.DocumentContent, userID string, options *FileParsingOptions) (*document.StructuredDocumentOutput, error)

	// RegisterStrategy adds a parsing strategy to the orchestrator
	// Strategies can be added at runtime to support new file types
	RegisterStrategy(strategy DocumentParsingStrategy) error

	// GetSupportedFileTypes returns all supported file MIME types
	GetSupportedFileTypes() []string

	// GetStrategyForFileType returns the appropriate strategy for a file type
	GetStrategyForFileType(mimeType string, userContext *document.UserContext) (DocumentParsingStrategy, error)

	// IsFileTypeSupported checks if a file type is supported
	IsFileTypeSupported(mimeType string) bool
}

// DocumentParsingStrategy defines the interface for a specific document parsing implementation
type DocumentParsingStrategy interface {
	// Process parses a document and returns structured output
	Process(ctx context.Context, content *document.DocumentContent, options *FileParsingOptions) (*document.StructuredDocumentOutput, error)

	// Name returns the name of the strategy
	Name() string

	// SupportedFileTypes returns a list of MIME types supported by this strategy
	SupportedFileTypes() []string

	// IsApplicable checks if this strategy can handle the given file type and user context
	// User context is used to determine if the user has access to premium features, etc.
	IsApplicable(ctx context.Context, mimeType string, userContext *document.UserContext) bool

	// Priority returns the priority of this strategy for the given file type
	// Higher priority strategies are tried first
	Priority(mimeType string) int

	// GetRequiredPreprocessing returns any preprocessing steps needed
	// For example, converting a docx to pdf before parsing
	GetRequiredPreprocessing(mimeType string) []string
}

// FileTypeConverterPort defines the interface for converting between file formats
type FileTypeConverterPort interface {
	// ConvertFile converts a file from one format to another
	// Returns the storage ID of the converted file
	ConvertFile(ctx context.Context, sourceID document.StorageObjectID,
		sourceMimeType, targetMimeType string) (document.StorageObjectID, error)

	// SupportsConversion checks if conversion between two formats is supported
	SupportsConversion(sourceMimeType, targetMimeType string) bool

	// GetSupportedSourceFormats returns all supported source formats
	GetSupportedSourceFormats() []string

	// GetSupportedTargetFormats returns all supported target formats for a source format
	GetSupportedTargetFormats(sourceMimeType string) []string
}
