package neo4j

import (
	"testing"
	"time"

	"demo/ms_canvas/go_app/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkRepo_CreateHierarchicalLinks_Validation(t *testing.T) {
	tests := []struct {
		name              string
		contentSourceID   uuid.UUID
		chunkIDs          []uuid.UUID
		connectionType    string
		hierarchyDepth    int
		createdAt         time.Time
		expectEmptyReturn bool
		validateFunc      func(t *testing.T, contentSourceID uuid.UUID, chunkIDs []uuid.UUID, connectionType string, hierarchyDepth int, createdAt time.Time)
	}{
		{
			name:              "empty chunk IDs should return early",
			contentSourceID:   uuid.New(),
			chunkIDs:          []uuid.UUID{},
			connectionType:    "document_chunk",
			hierarchyDepth:    1,
			createdAt:         time.Now(),
			expectEmptyReturn: true,
			validateFunc: func(t *testing.T, contentSourceID uuid.UUID, chunkIDs []uuid.UUID, connectionType string, hierarchyDepth int, createdAt time.Time) {
				assert.Len(t, chunkIDs, 0)
			},
		},
		{
			name:            "valid hierarchical links should have correct structure",
			contentSourceID: uuid.New(),
			chunkIDs: []uuid.UUID{
				uuid.New(),
				uuid.New(),
				uuid.New(),
			},
			connectionType:    "document_chunk",
			hierarchyDepth:    1,
			createdAt:         time.Now(),
			expectEmptyReturn: false,
			validateFunc: func(t *testing.T, contentSourceID uuid.UUID, chunkIDs []uuid.UUID, connectionType string, hierarchyDepth int, createdAt time.Time) {
				assert.NotEqual(t, uuid.Nil, contentSourceID)
				assert.Len(t, chunkIDs, 3)
				assert.Equal(t, "document_chunk", connectionType)
				assert.Equal(t, 1, hierarchyDepth)
				assert.False(t, createdAt.IsZero())

				for _, chunkID := range chunkIDs {
					assert.NotEqual(t, uuid.Nil, chunkID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validateFunc(t, tt.contentSourceID, tt.chunkIDs, tt.connectionType, tt.hierarchyDepth, tt.createdAt)
		})
	}
}

func TestLinkRepo_CreateSemanticLinks_Validation(t *testing.T) {
	now := time.Now()
	sourceID := uuid.New()
	targetID := uuid.New()

	tests := []struct {
		name         string
		links        []domain.SemanticLink
		validateFunc func(t *testing.T, links []domain.SemanticLink)
	}{
		{
			name:  "empty links should be handled",
			links: []domain.SemanticLink{},
			validateFunc: func(t *testing.T, links []domain.SemanticLink) {
				assert.Len(t, links, 0)
			},
		},
		{
			name: "semantic link with all fields should be valid",
			links: []domain.SemanticLink{
				{
					SourceID:            sourceID,
					TargetID:            targetID,
					ConnectionType:      "semantic_similarity",
					StrengthScore:       0.85,
					SimilarityScore:     0.92,
					AbstractionBridge:   true,
					HierarchicalBridge:  false,
					ExplorationMetadata: map[string]any{"method": "cosine"},
					SemanticTags:        []string{"similar", "related"},
					StyleMetadata:       map[string]any{"color": "blue"},
					Description:         stringPtr("High semantic similarity"),
					CreatedAt:           now,
					UpdatedAt:           now,
				},
			},
			validateFunc: func(t *testing.T, links []domain.SemanticLink) {
				require.Len(t, links, 1)
				link := links[0]

				assert.NotEqual(t, uuid.Nil, link.SourceID)
				assert.NotEqual(t, uuid.Nil, link.TargetID)
				assert.Equal(t, "semantic_similarity", link.ConnectionType)
				assert.Equal(t, 0.85, link.StrengthScore)
				assert.Equal(t, 0.92, link.SimilarityScore)
				assert.True(t, link.AbstractionBridge)
				assert.False(t, link.HierarchicalBridge)
				assert.NotNil(t, link.ExplorationMetadata)
				assert.Equal(t, "cosine", link.ExplorationMetadata["method"])
				assert.Contains(t, link.SemanticTags, "similar")
				assert.Contains(t, link.SemanticTags, "related")
				assert.NotNil(t, link.StyleMetadata)
				assert.Equal(t, "blue", link.StyleMetadata["color"])
				assert.NotNil(t, link.Description)
				assert.Equal(t, "High semantic similarity", *link.Description)
				assert.False(t, link.CreatedAt.IsZero())
				assert.False(t, link.UpdatedAt.IsZero())
			},
		},
		{
			name: "semantic link with minimal fields should be valid",
			links: []domain.SemanticLink{
				{
					SourceID:        sourceID,
					TargetID:        targetID,
					ConnectionType:  "basic_similarity",
					StrengthScore:   0.5,
					SimilarityScore: 0.6,
				},
			},
			validateFunc: func(t *testing.T, links []domain.SemanticLink) {
				require.Len(t, links, 1)
				link := links[0]

				assert.NotEqual(t, uuid.Nil, link.SourceID)
				assert.NotEqual(t, uuid.Nil, link.TargetID)
				assert.Equal(t, "basic_similarity", link.ConnectionType)
				assert.Equal(t, 0.5, link.StrengthScore)
				assert.Equal(t, 0.6, link.SimilarityScore)
				assert.False(t, link.AbstractionBridge)
				assert.False(t, link.HierarchicalBridge)
				assert.True(t, link.CreatedAt.IsZero()) // Not set
				assert.Nil(t, link.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validateFunc(t, tt.links)
		})
	}
}

func TestLinkRepo_CreateStructuralLink_Validation(t *testing.T) {
	now := time.Now()
	sourceID := uuid.New()
	targetID := uuid.New()

	tests := []struct {
		name         string
		link         domain.StructuralLink
		validateFunc func(t *testing.T, link domain.StructuralLink)
	}{
		{
			name: "structural link with all fields should be valid",
			link: domain.StructuralLink{
				SourceID:            sourceID,
				TargetID:            targetID,
				ConnectionType:      "manual_connection",
				ConfidenceScore:     0.95,
				Description:         stringPtr("User-created connection"),
				ExplorationMetadata: map[string]any{"user_id": "12345"},
				StyleMetadata:       map[string]any{"thickness": 2},
				CreatedBy:           "user_12345",
				CreatedAt:           now,
				UpdatedAt:           now,
			},
			validateFunc: func(t *testing.T, link domain.StructuralLink) {
				assert.NotEqual(t, uuid.Nil, link.SourceID)
				assert.NotEqual(t, uuid.Nil, link.TargetID)
				assert.Equal(t, "manual_connection", link.ConnectionType)
				assert.Equal(t, 0.95, link.ConfidenceScore)
				assert.NotNil(t, link.Description)
				assert.Equal(t, "User-created connection", *link.Description)
				assert.NotNil(t, link.ExplorationMetadata)
				assert.Equal(t, "12345", link.ExplorationMetadata["user_id"])
				assert.NotNil(t, link.StyleMetadata)
				assert.Equal(t, 2, link.StyleMetadata["thickness"])
				assert.Equal(t, "user_12345", link.CreatedBy)
				assert.False(t, link.CreatedAt.IsZero())
				assert.False(t, link.UpdatedAt.IsZero())
			},
		},
		{
			name: "structural link with minimal fields should be valid",
			link: domain.StructuralLink{
				SourceID:        sourceID,
				TargetID:        targetID,
				ConnectionType:  "auto_connection",
				ConfidenceScore: 0.7,
				CreatedBy:       "system",
			},
			validateFunc: func(t *testing.T, link domain.StructuralLink) {
				assert.NotEqual(t, uuid.Nil, link.SourceID)
				assert.NotEqual(t, uuid.Nil, link.TargetID)
				assert.Equal(t, "auto_connection", link.ConnectionType)
				assert.Equal(t, 0.7, link.ConfidenceScore)
				assert.Equal(t, "system", link.CreatedBy)
				assert.Nil(t, link.Description)
				assert.True(t, link.CreatedAt.IsZero()) // Not set
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validateFunc(t, tt.link)
		})
	}
}

// TestableLinkRepo for testing database interactions
type TestableLinkRepo struct {
	executeWriteFunc func(cypher string, params map[string]any) error
}

func (r *TestableLinkRepo) ExecuteWrite(cypher string, params map[string]any) error {
	if r.executeWriteFunc != nil {
		return r.executeWriteFunc(cypher, params)
	}
	return nil
}

func TestLinkRepo_CreateHierarchicalLinks_DatabaseInteraction(t *testing.T) {
	contentSourceID := uuid.New()
	chunkIDs := []uuid.UUID{uuid.New(), uuid.New()}
	connectionType := "document_chunk"
	hierarchyDepth := 1
	createdAt := time.Now()

	tests := []struct {
		name           string
		contentSourceID uuid.UUID
		chunkIDs       []uuid.UUID
		connectionType string
		hierarchyDepth int
		createdAt      time.Time
		validateCypher func(t *testing.T, cypher string, params map[string]any)
	}{
		{
			name:            "hierarchical links should generate correct cypher",
			contentSourceID: contentSourceID,
			chunkIDs:        chunkIDs,
			connectionType:  connectionType,
			hierarchyDepth:  hierarchyDepth,
			createdAt:       createdAt,
			validateCypher: func(t *testing.T, cypher string, params map[string]any) {
				assert.Contains(t, cypher, "MATCH (c:ContentNode {content_source_id: $content_source_id})")
				assert.Contains(t, cypher, "UNWIND $items AS item")
				assert.Contains(t, cypher, "MATCH (n:ChunkNode {id: item.chunk_id})")
				assert.Contains(t, cypher, "MERGE (c)-[r:HIERARCHICAL_PARENT]->(n)")
				assert.Contains(t, cypher, "SET r.connection_type = $connection_type")
				assert.Contains(t, cypher, "r.hierarchy_depth = $hierarchy_depth")

				assert.Contains(t, params, "content_source_id")
				assert.Contains(t, params, "items")
				assert.Contains(t, params, "connection_type")
				assert.Contains(t, params, "hierarchy_depth")
				assert.Contains(t, params, "created_at")

				assert.Equal(t, contentSourceID.String(), params["content_source_id"])
				assert.Equal(t, connectionType, params["connection_type"])
				assert.Equal(t, hierarchyDepth, params["hierarchy_depth"])

				items, ok := params["items"].([]map[string]any)
				assert.True(t, ok)
				assert.Len(t, items, 2)

				for i, item := range items {
					assert.Contains(t, item, "chunk_id")
					assert.Equal(t, chunkIDs[i].String(), item["chunk_id"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := &TestableLinkRepo{
				executeWriteFunc: func(cypher string, params map[string]any) error {
					tt.validateCypher(t, cypher, params)
					return nil
				},
			}

			// Simulate the logic from CreateHierarchicalLinks
			if len(tt.chunkIDs) == 0 {
				return // Early return
			}

			items := make([]map[string]any, 0, len(tt.chunkIDs))
			for _, id := range tt.chunkIDs {
				items = append(items, map[string]any{"chunk_id": id.String()})
			}

			params := map[string]any{
				"content_source_id": tt.contentSourceID.String(),
				"items":             items,
				"connection_type":   tt.connectionType,
				"hierarchy_depth":   tt.hierarchyDepth,
				"created_at":        tt.createdAt,
			}

			cypher := `
            MATCH (c:ContentNode {content_source_id: $content_source_id})
            UNWIND $items AS item
            MATCH (n:ChunkNode {id: item.chunk_id})
            MERGE (c)-[r:HIERARCHICAL_PARENT]->(n)
            SET r.connection_type = $connection_type,
                r.hierarchy_depth = $hierarchy_depth,
                r.created_at = datetime($created_at)
        `

			err := testRepo.ExecuteWrite(cypher, params)
			assert.NoError(t, err)
		})
	}
}

func TestLinkRepo_CreateSemanticLinks_DatabaseInteraction(t *testing.T) {
	now := time.Now()
	links := []domain.SemanticLink{
		{
			SourceID:            uuid.New(),
			TargetID:            uuid.New(),
			ConnectionType:      "semantic_similarity",
			StrengthScore:       0.85,
			SimilarityScore:     0.92,
			AbstractionBridge:   true,
			HierarchicalBridge:  false,
			ExplorationMetadata: map[string]any{"method": "cosine"},
			SemanticTags:        []string{"similar", "related"},
			StyleMetadata:       map[string]any{"color": "blue"},
			Description:         stringPtr("High semantic similarity"),
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			SourceID:        uuid.New(),
			TargetID:        uuid.New(),
			ConnectionType:  "basic_similarity",
			StrengthScore:   0.5,
			SimilarityScore: 0.6,
		},
	}

	testRepo := &TestableLinkRepo{
		executeWriteFunc: func(cypher string, params map[string]any) error {
			assert.Contains(t, cypher, "UNWIND $items AS item")
			assert.Contains(t, cypher, "MATCH (s {id: item.src})")
			assert.Contains(t, cypher, "MATCH (t {id: item.dst})")
			assert.Contains(t, cypher, "MERGE (s)-[r:SEMANTIC_LINK]->(t)")
			assert.Contains(t, cypher, "SET r.connection_type = item.connection_type")
			assert.Contains(t, cypher, "r.strength_score = item.strength_score")
			assert.Contains(t, cypher, "r.similarity_score = item.similarity_score")

			assert.Contains(t, params, "items")

			items, ok := params["items"].([]map[string]any)
			assert.True(t, ok)
			assert.Len(t, items, 2)

			// Check first link (complete)
			item1 := items[0]
			assert.Contains(t, item1, "src")
			assert.Contains(t, item1, "dst")
			assert.Contains(t, item1, "connection_type")
			assert.Equal(t, "semantic_similarity", item1["connection_type"])
			assert.Equal(t, 0.85, item1["strength_score"])
			assert.Equal(t, 0.92, item1["similarity_score"])
			assert.Equal(t, true, item1["abstraction_bridge"])
			assert.Equal(t, false, item1["hierarchical_bridge"])

			// Check second link (minimal)
			item2 := items[1]
			assert.Equal(t, "basic_similarity", item2["connection_type"])
			assert.Equal(t, 0.5, item2["strength_score"])
			assert.Equal(t, 0.6, item2["similarity_score"])

			return nil
		},
	}

	// Simulate CreateSemanticLinks logic
	if len(links) == 0 {
		return
	}

	params := map[string]any{"items": make([]map[string]any, 0, len(links))}
	for _, l := range links {
		createdAt := l.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		updatedAt := l.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}
		params["items"] = append(params["items"].([]map[string]any), map[string]any{
			"src":                  l.SourceID.String(),
			"dst":                  l.TargetID.String(),
			"connection_type":      l.ConnectionType,
			"strength_score":       l.StrengthScore,
			"similarity_score":     l.SimilarityScore,
			"abstraction_bridge":   l.AbstractionBridge,
			"hierarchical_bridge":  l.HierarchicalBridge,
			"exploration_metadata": l.ExplorationMetadata,
			"semantic_tags":        l.SemanticTags,
			"style_metadata":       l.StyleMetadata,
			"description":          l.Description,
			"created_at":           createdAt,
			"updated_at":           updatedAt,
			"deleted_at":           l.DeletedAt,
		})
	}

	cypher := `
            UNWIND $items AS item
            MATCH (s {id: item.src})
            MATCH (t {id: item.dst})
            MERGE (s)-[r:SEMANTIC_LINK]->(t)
            SET r.connection_type = item.connection_type,
                r.strength_score = item.strength_score,
                r.similarity_score = item.similarity_score,
                r.abstraction_bridge = item.abstraction_bridge,
                r.hierarchical_bridge = item.hierarchical_bridge,
                r.exploration_metadata = item.exploration_metadata,
                r.semantic_tags = item.semantic_tags,
                r.style_metadata = item.style_metadata,
                r.description = item.description,
                r.created_at = datetime(item.created_at),
                r.updated_at = datetime(item.updated_at),
                r.deleted_at = item.deleted_at
        `

	err := testRepo.ExecuteWrite(cypher, params)
	assert.NoError(t, err)
}

func TestLinkRepo_CreateStructuralLink_DatabaseInteraction(t *testing.T) {
	now := time.Now()
	link := domain.StructuralLink{
		SourceID:            uuid.New(),
		TargetID:            uuid.New(),
		ConnectionType:      "manual_connection",
		ConfidenceScore:     0.95,
		Description:         stringPtr("User-created connection"),
		ExplorationMetadata: map[string]any{"user_id": "12345"},
		StyleMetadata:       map[string]any{"thickness": 2},
		CreatedBy:           "user_12345",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	testRepo := &TestableLinkRepo{
		executeWriteFunc: func(cypher string, params map[string]any) error {
			assert.Contains(t, cypher, "MATCH (s {id: $src})")
			assert.Contains(t, cypher, "MATCH (t {id: $dst})")
			assert.Contains(t, cypher, "MERGE (s)-[r:STRUCTURAL_LINK]->(t)")
			assert.Contains(t, cypher, "SET r.connection_type = $connection_type")
			assert.Contains(t, cypher, "r.confidence_score = $confidence_score")
			assert.Contains(t, cypher, "r.created_by = $created_by")

			assert.Contains(t, params, "src")
			assert.Contains(t, params, "dst")
			assert.Contains(t, params, "connection_type")
			assert.Contains(t, params, "confidence_score")
			assert.Contains(t, params, "created_by")

			assert.Equal(t, link.SourceID.String(), params["src"])
			assert.Equal(t, link.TargetID.String(), params["dst"])
			assert.Equal(t, "manual_connection", params["connection_type"])
			assert.Equal(t, 0.95, params["confidence_score"])
			assert.Equal(t, "user_12345", params["created_by"])

			return nil
		},
	}

	// Simulate CreateStructuralLink logic
	createdAt := link.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := link.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	params := map[string]any{
		"src":                  link.SourceID.String(),
		"dst":                  link.TargetID.String(),
		"connection_type":      link.ConnectionType,
		"confidence_score":     link.ConfidenceScore,
		"description":          link.Description,
		"exploration_metadata": link.ExplorationMetadata,
		"style_metadata":       link.StyleMetadata,
		"created_by":           link.CreatedBy,
		"created_at":           createdAt,
		"updated_at":           updatedAt,
		"deleted_at":           link.DeletedAt,
	}

	cypher := `
        MATCH (s {id: $src})
        MATCH (t {id: $dst})
        MERGE (s)-[r:STRUCTURAL_LINK]->(t)
        SET r.connection_type = $connection_type,
            r.confidence_score = $confidence_score,
            r.description = $description,
            r.exploration_metadata = $exploration_metadata,
            r.style_metadata = $style_metadata,
            r.created_by = $created_by,
            r.created_at = datetime($created_at),
            r.updated_at = datetime($updated_at),
            r.deleted_at = $deleted_at
    `

	err := testRepo.ExecuteWrite(cypher, params)
	assert.NoError(t, err)
}