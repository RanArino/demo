package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	v1 "demo/ms_canvas/go_app/api/proto/public/v1"
	"demo/ms_canvas/go_app/internal/repository/neo4j"

	"github.com/google/uuid"
)

// SearchService handles semantic search operations

// searchServiceImpl implements SearchService
type searchServiceImpl struct {
	searchRepo *neo4j.SearchRepo
}

// NewSearchService creates a new SearchService with dependencies
func NewSearchService(searchRepo *neo4j.SearchRepo) SearchService {
	return &searchServiceImpl{
		searchRepo: searchRepo,
	}
}

func (s *searchServiceImpl) SemanticSearch(ctx context.Context, req *v1.SemanticSearchRequest) (*v1.SemanticSearchResponse, error) {
	// Input validation
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 25 // Default value
	}

	if topK > 100 {
		topK = 100 // Maximum value to prevent abuse
	}

	// Parse space ID
	spaceUUID, err := uuid.Parse(req.SpaceId)
	if err != nil {
		return nil, fmt.Errorf("invalid space ID: %w", err)
	}

	log.Printf("Performing semantic search for query: '%s' in space: '%s' with topK: %d", req.Query, req.SpaceId, topK)

	// Search for nodes in the specified space using vector search
	// Embedding is generated in-database using Neo4j's genai.vector.encode
	nodes, scores, err := s.searchRepo.VectorSearch(ctx, spaceUUID, req.Query, topK, req.NodeTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to perform vector search: %w", err)
	}

	// Convert nodes to search results with actual scores from vector search
	results := make([]*v1.SearchResult, 0, len(nodes))
	for i, node := range nodes {
		var score float64
		if i < len(scores) {
			score = scores[i]
		}

		result := &v1.SearchResult{
			Node:  node,
			Score: score,
		}

		results = append(results, result)
	}

	log.Printf("Search completed, returning %d results", len(results))

	// Return the proper response format
	return &v1.SemanticSearchResponse{
		Results: results,
	}, nil
}

// calculateSimilarityScore calculates a simple similarity score
// In production, this would use vector similarity calculations
func (s *searchServiceImpl) calculateSimilarityScore(query string, node *v1.Node) float64 {
	// Simple text-based similarity for now
	// In production, this would use cosine similarity or other vector metrics

	queryLower := strings.ToLower(query)

	switch n := node.GetNode().(type) {
	case *v1.Node_Content:
		contentNode := n.Content
		if contentNode.Base != nil && contentNode.Base.DisplayContent != nil {
			content := strings.ToLower(*contentNode.Base.DisplayContent)

			// Simple term matching score
			terms := strings.Fields(queryLower)
			matches := 0
			for _, term := range terms {
				if strings.Contains(content, term) {
					matches++
				}
			}

			if len(terms) > 0 {
				return float64(matches) / float64(len(terms))
			}
		}
	case *v1.Node_Chunk:
		chunkNode := n.Chunk
		var content string
		if chunkNode.Base != nil && chunkNode.Base.ChatContent != nil {
			content = strings.ToLower(*chunkNode.Base.ChatContent)
		} else {
			content = ""
		}

		terms := strings.Fields(queryLower)
		matches := 0
		for _, term := range terms {
			if strings.Contains(content, term) {
				matches++
			}
		}

		if len(terms) > 0 {
			return float64(matches) / float64(len(terms))
		}
	case *v1.Node_Cluster:
		clusterNode := n.Cluster
		if clusterNode.Base != nil && clusterNode.Base.DisplayContent != nil {
			content := strings.ToLower(*clusterNode.Base.DisplayContent)

			terms := strings.Fields(queryLower)
			matches := 0
			for _, term := range terms {
				if strings.Contains(content, term) {
					matches++
				}
			}

			if len(terms) > 0 {
				return float64(matches) / float64(len(terms))
			}
		}
	}

	return 0.0 // No similarity
}
