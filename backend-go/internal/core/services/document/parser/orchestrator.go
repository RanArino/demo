package parser

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"

	"github.com/ran/demo/backend-go/internal/core/models/document"
	ports "github.com/ran/demo/backend-go/internal/core/ports/document"
)

// FileParsingOrchestrator implements the FileParsingOrchestratorPort interface
type FileParsingOrchestrator struct {
	strategies map[string]ports.DocumentParsingStrategy
	mutex      sync.RWMutex
	storage    ports.StoragePort
	converter  ports.FileTypeConverterPort
	logger     *slog.Logger
}

// NewFileParsingOrchestrator creates a new file parsing orchestrator
func NewFileParsingOrchestrator(
	storage ports.StoragePort,
	converter ports.FileTypeConverterPort,
	logger *slog.Logger,
) *FileParsingOrchestrator {
	return &FileParsingOrchestrator{
		strategies: make(map[string]ports.DocumentParsingStrategy),
		storage:    storage,
		converter:  converter,
		logger:     logger,
	}
}

// ParseDocument parses a document using the appropriate strategy
func (o *FileParsingOrchestrator) ParseDocument(
	ctx context.Context,
	content *document.DocumentContent,
	userID string,
	options *ports.FileParsingOptions,
) (*document.StructuredDocumentOutput, error) {
	// Create user context
	userContext := &document.UserContext{
		UserID: userID,
		Tier:   "standard", // This would come from user service in a real implementation
		Domain: "general",  // This would come from user service in a real implementation
	}

	// Get strategy for the file type
	strategy, err := o.GetStrategyForFileType(content.MIMEType, userContext)
	if err != nil {
		return nil, err
	}

	// Check if preprocessing is needed
	preprocessSteps := strategy.GetRequiredPreprocessing(content.MIMEType)
	if len(preprocessSteps) > 0 {
		// In a real implementation, we would perform preprocessing here
		// For simplicity, we'll just log that preprocessing is needed
		o.logger.Info("Preprocessing needed for document",
			"user_id", userID,
			"mime_type", content.MIMEType,
			"preprocessing_steps", preprocessSteps)
	}

	// If a specific strategy was requested, override the default selection
	if options != nil && options.PreferredStrategy != "" {
		o.mutex.RLock()
		preferredStrategy, exists := o.strategies[options.PreferredStrategy]
		o.mutex.RUnlock()

		if exists && preferredStrategy.IsApplicable(ctx, content.MIMEType, userContext) {
			strategy = preferredStrategy
		} else {
			o.logger.Info("Requested strategy not available or applicable, falling back to default",
				"requested_strategy", options.PreferredStrategy,
				"mime_type", content.MIMEType)
		}
	}

	// Set default options if none were provided
	if options == nil {
		options = &ports.FileParsingOptions{
			ExtractTables: true,
			ExtractForms:  true,
			ExtractImages: true,
			Language:      "en",
		}
	}

	// Log parsing start
	o.logger.Info("Starting document parsing",
		"strategy", strategy.Name(),
		"mime_type", content.MIMEType,
		"extract_tables", options.ExtractTables,
		"extract_forms", options.ExtractForms,
		"extract_images", options.ExtractImages)

	// Parse the document
	output, err := strategy.Process(ctx, content, options)
	if err != nil {
		o.logger.Error("Document parsing failed",
			"error", err,
			"strategy", strategy.Name(),
			"mime_type", content.MIMEType)
		return nil, err
	}

	// Log parsing completion
	o.logger.Info("Document parsing completed",
		"strategy", strategy.Name(),
		"mime_type", content.MIMEType)

	return output, nil
}

// RegisterStrategy adds a parsing strategy to the orchestrator
func (o *FileParsingOrchestrator) RegisterStrategy(strategy ports.DocumentParsingStrategy) error {
	if strategy == nil {
		return fmt.Errorf("strategy cannot be nil")
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.strategies[strategy.Name()] = strategy
	o.logger.Info("Strategy registered", "name", strategy.Name())
	return nil
}

// GetSupportedFileTypes returns all supported file MIME types
func (o *FileParsingOrchestrator) GetSupportedFileTypes() []string {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Get all unique MIME types supported by all strategies
	mimeTypes := make(map[string]struct{})
	for _, strategy := range o.strategies {
		for _, mimeType := range strategy.SupportedFileTypes() {
			mimeTypes[mimeType] = struct{}{}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(mimeTypes))
	for mimeType := range mimeTypes {
		result = append(result, mimeType)
	}

	// Sort for consistent results
	sort.Strings(result)
	return result
}

// GetStrategyForFileType returns the appropriate strategy for a file type
func (o *FileParsingOrchestrator) GetStrategyForFileType(
	mimeType string,
	userContext *document.UserContext,
) (ports.DocumentParsingStrategy, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	ctx := context.Background()

	var applicableStrategies []ports.DocumentParsingStrategy
	for _, strategy := range o.strategies {
		if strategy.IsApplicable(ctx, mimeType, userContext) {
			applicableStrategies = append(applicableStrategies, strategy)
		}
	}

	if len(applicableStrategies) == 0 {
		return nil, fmt.Errorf("no strategy available for MIME type: %s", mimeType)
	}

	// Get the highest priority strategy
	sort.Slice(applicableStrategies, func(i, j int) bool {
		return applicableStrategies[i].Priority(mimeType) > applicableStrategies[j].Priority(mimeType)
	})

	return applicableStrategies[0], nil
}

// IsFileTypeSupported checks if a file type is supported
func (o *FileParsingOrchestrator) IsFileTypeSupported(mimeType string) bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, strategy := range o.strategies {
		for _, supportedType := range strategy.SupportedFileTypes() {
			if supportedType == mimeType {
				return true
			}
		}
	}

	return false
}
