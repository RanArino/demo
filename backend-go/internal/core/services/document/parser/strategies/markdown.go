package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"regexp"

	"github.com/google/uuid"
	"github.com/ran/demo/backend-go/internal/core/models/document"
	ports "github.com/ran/demo/backend-go/internal/core/ports/document"
)

// MarkdownStrategy implements a parsing strategy for Markdown files
type MarkdownStrategy struct {
	storage ports.StoragePort
	logger  *slog.Logger
}

// NewMarkdownStrategy creates a new Markdown parsing strategy
func NewMarkdownStrategy(storage ports.StoragePort, logger *slog.Logger) *MarkdownStrategy {
	return &MarkdownStrategy{
		storage: storage,
		logger:  logger,
	}
}

// Name returns the name of the strategy
func (s *MarkdownStrategy) Name() string {
	return "markdown"
}

// SupportedFileTypes returns a list of MIME types supported by this strategy
func (s *MarkdownStrategy) SupportedFileTypes() []string {
	return []string{
		"text/markdown",
		"text/x-markdown",
		"text/plain", // Treat plain text as markdown for simplicity
	}
}

// IsApplicable checks if this strategy can handle the given file type and user context
func (s *MarkdownStrategy) IsApplicable(
	ctx context.Context,
	mimeType string,
	userContext *document.UserContext,
) bool {
	for _, supportedType := range s.SupportedFileTypes() {
		if supportedType == mimeType {
			return true
		}
	}
	return false
}

// Priority returns the priority of this strategy for the given file type
func (s *MarkdownStrategy) Priority(mimeType string) int {
	switch mimeType {
	case "text/markdown", "text/x-markdown":
		return 100 // High priority for markdown files
	case "text/plain":
		return 50 // Medium priority for text files
	default:
		return 0
	}
}

// GetRequiredPreprocessing returns any preprocessing steps needed
func (s *MarkdownStrategy) GetRequiredPreprocessing(mimeType string) []string {
	// No preprocessing needed for markdown
	return []string{}
}

// Process parses a document and returns structured output
func (s *MarkdownStrategy) Process(
	ctx context.Context,
	content *document.DocumentContent,
	options *ports.FileParsingOptions,
) (*document.StructuredDocumentOutput, error) {
	// Read content
	data, err := io.ReadAll(content.Stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown content: %w", err)
	}

	// Parse markdown content
	markdownContent := string(data)

	// Extract title from first heading
	title := extractTitle(markdownContent)

	// Extract tables if requested
	var tables []document.TableData
	if options.ExtractTables {
		tables = extractTables(markdownContent, string(options.Language))
	}

	// Create JSON for structured data
	tablesJSON, err := json.Marshal(tables)
	if err != nil {
		s.logger.Error("Failed to marshal tables to JSON", "error", err)
		tablesJSON = []byte("[]")
	}

	// Create structured output
	output := &document.StructuredDocumentOutput{
		MarkdownContent: markdownContent,
		RawText:         markdownContent, // For markdown, raw text is the same as markdown content
		Title:           title,
		Tables:          tables,
		TablesJSON:      string(tablesJSON),
		ImagesJSON:      "[]", // No images in plain markdown
		MetadataJSON:    "{}", // No special metadata for markdown
	}

	return output, nil
}

// extractTitle extracts the title from markdown content (first # heading)
func extractTitle(content string) string {
	// Look for first heading
	re := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		return matches[1]
	}

	// If no heading, return empty string
	return ""
}

// extractTables extracts tables from markdown content
func extractTables(content string, language string) []document.TableData {
	// Simple table regex for markdown tables
	tableRegex := regexp.MustCompile(`(?ms)\|(.+)\|\s*\n\|([\s-:|]+)\|\s*\n((?:\|.+\|\s*\n)+)`)
	matches := tableRegex.FindAllStringSubmatch(content, -1)

	var tables []document.TableData

	for i, match := range matches {
		if len(match) < 4 {
			continue
		}

		// Extract headers
		headerRow := match[1]
		headerRegex := regexp.MustCompile(`\s*([^|]+)\s*`)
		headerMatches := headerRegex.FindAllStringSubmatch(headerRow, -1)

		headers := make([]string, 0, len(headerMatches))
		for _, header := range headerMatches {
			if len(header) > 1 {
				headers = append(headers, header[1])
			}
		}

		// Extract rows
		rowsContent := match[3]
		rowRegex := regexp.MustCompile(`(?m)^\|(.*)\|$`)
		rowMatches := rowRegex.FindAllStringSubmatch(rowsContent, -1)

		rows := make([][]string, 0, len(rowMatches))
		for _, rowMatch := range rowMatches {
			if len(rowMatch) < 2 {
				continue
			}

			cellRegex := regexp.MustCompile(`\s*([^|]+)\s*`)
			cellMatches := cellRegex.FindAllStringSubmatch(rowMatch[1], -1)

			row := make([]string, 0, len(cellMatches))
			for _, cell := range cellMatches {
				if len(cell) > 1 {
					row = append(row, cell[1])
				}
			}

			rows = append(rows, row)
		}

		// Create table data
		table := document.TableData{
			ID:          uuid.New().String(),
			Headers:     headers,
			Rows:        rows,
			Description: fmt.Sprintf("Table %d", i+1),
		}

		tables = append(tables, table)
	}

	return tables
}
