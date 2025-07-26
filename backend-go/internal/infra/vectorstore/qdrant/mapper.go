package qdrant

import (
	"fmt"

	"github.com/ran/demo/backend-go/internal/domain/models"
)

// Point represents a Qdrant point payload in our domain
// ID is the unique point identifier; Payload contains vector metadata
type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// DocumentMeta represents the top-level document info for Qdrant
type DocumentMeta struct {
	DocumentId       string    `json:"documentId"`
	FileName         string    `json:"fileName"`
	SummaryText      string    `json:"summaryText"`
	SummaryPosition  []float64 `json:"summaryPosition"`
	SummaryClusterId string    `json:"summaryClusterId"`
	ChunkIds         []string  `json:"chunkIds"`
}

// MapToQdrantPoints merges chunks with embeddings and coords into Point structs, including cluster IDs and keywords
func MapToQdrantPoints(chunks []models.Chunk, embeddings [][]float32, coords [][]float64) ([]Point, error) {
	n := len(chunks)
	if len(embeddings) != n || len(coords) != n {
		return nil, fmt.Errorf("length mismatch: chunks=%d, embeddings=%d, coords=%d", n, len(embeddings), len(coords))
	}
	points := make([]Point, n)
	for i, c := range chunks {
		points[i] = Point{
			ID:     c.ID,
			Vector: embeddings[i],
			Payload: map[string]interface{}{
				"documentId": c.DocumentID,
				"chunkId":    c.ID,
				"text":       c.Text,
				"position":   coords[i],
				"clusterIds": c.ClusterIDs,
				"keywords":   c.Keywords,
			},
		}
	}
	return points, nil
}

// MapDocMeta creates DocumentMeta for overall document summary
func MapDocMeta(doc models.Document, summary models.Summary, chunkIDs []string, summaryPosition []float64, summaryClusterId string) (DocumentMeta, error) {
	if err := doc.Validate(); err != nil {
		return DocumentMeta{}, fmt.Errorf("invalid document %s: %w", doc.ID, err)
	}
	if err := summary.Validate(); err != nil {
		return DocumentMeta{}, fmt.Errorf("invalid summary %s: %w", summary.ID, err)
	}
	return DocumentMeta{
		DocumentId:       doc.ID,
		FileName:         doc.Filename,
		SummaryText:      summary.Text,
		SummaryPosition:  summaryPosition,
		SummaryClusterId: summaryClusterId,
		ChunkIds:         chunkIDs,
	}, nil
}
