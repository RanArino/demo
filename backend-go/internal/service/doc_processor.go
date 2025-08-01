package service

import (
	"fmt"

	"github.com/ran/demo/backend-go/internal/domain/models"
)

// BuildChunks wraps each text into a Chunk, assigning IDs, token counts, and keywords.
func BuildChunks(docID string, texts []string, keywords [][]string) ([]models.Chunk, error) {
	if len(keywords) != len(texts) {
		return nil, fmt.Errorf("keywords length %d does not match texts length %d", len(keywords), len(texts))
	}
	var chunks []models.Chunk
	for i, txt := range texts {
		id := fmt.Sprintf("%s_%d", docID, i)
		tokenCount := CountTokens(txt)
		chunk := models.Chunk{
			ID:         id,
			DocumentID: docID,
			Index:      i,
			Text:       txt,
			TokenCount: tokenCount,
			Keywords:   keywords[i],
		}
		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk %s: %w", id, err)
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// BuildSummary wraps summaryText into a Summary model for the document.
func BuildSummary(docID, summaryText string) (models.Summary, error) {
	id := fmt.Sprintf("%s_summary", docID)
	summary := models.Summary{
		ID:         id,
		DocumentID: docID,
		Text:       summaryText,
	}
	if err := summary.Validate(); err != nil {
		return models.Summary{}, fmt.Errorf("invalid summary %s: %w", id, err)
	}
	return summary, nil
}