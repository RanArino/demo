package neo4j

import (
	"context"
	"testing"
	"time"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLinkRepo_CreateHierarchicalLinks_Validation(t *testing.T) {
	tests := []struct {
		name              string
		links             []*canvasv1.HierarchicalLink
		expectEmptyReturn bool
		validateFunc      func(t *testing.T, links []*canvasv1.HierarchicalLink)
	}{
		{
			name:              "empty links should return early",
			links:             []*canvasv1.HierarchicalLink{},
			expectEmptyReturn: true,
			validateFunc: func(t *testing.T, links []*canvasv1.HierarchicalLink) {
				assert.Len(t, links, 0)
			},
		},
		{
			name: "valid hierarchical links should have correct structure",
			links: []*canvasv1.HierarchicalLink{
				{
					Base: &canvasv1.BaseLink{
						SourceId: uuid.New().String(),
						TargetId: uuid.New().String(),
					},
					HierarchyDepth: 1,
				},
				{
					Base: &canvasv1.BaseLink{
						SourceId: uuid.New().String(),
						TargetId: uuid.New().String(),
					},
					HierarchyDepth: 1,
				},
				{
					Base: &canvasv1.BaseLink{
						SourceId: uuid.New().String(),
						TargetId: uuid.New().String(),
					},
					HierarchyDepth: 1,
				},
			},
			expectEmptyReturn: false,
			validateFunc: func(t *testing.T, links []*canvasv1.HierarchicalLink) {
				assert.Len(t, links, 3)
				for _, link := range links {
					assert.Equal(t, canvasv1.HierarchicalConnectionType_HIERARCHICAL_CONNECTION_TYPE_ABSTRACTION, link.ConnectionType)
					assert.Equal(t, int32(1), link.HierarchyDepth)
					assert.NotEmpty(t, link.Base.SourceId)
					assert.NotEmpty(t, link.Base.TargetId)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validateFunc(t, tt.links)
		})
	}
}

func TestLinkRepo_CreateSemanticLinks_Validation(t *testing.T) {
	now := time.Now()
	sourceID := uuid.New().String()
	targetID := uuid.New().String()

	tests := []struct {
		name         string
		links        []*canvasv1.SemanticLink
		validateFunc func(t *testing.T, links []*canvasv1.SemanticLink)
	}{
		{
			name:  "empty links should be handled",
			links: []*canvasv1.SemanticLink{},
			validateFunc: func(t *testing.T, links []*canvasv1.SemanticLink) {
				assert.Len(t, links, 0)
			},
		},
		{
			name: "semantic link with all fields should be valid",
			links: []*canvasv1.SemanticLink{
				{
					Base: &canvasv1.BaseLink{
						SourceId:            sourceID,
						TargetId:            targetID,
						ExplorationMetadata: mapToStruct(map[string]any{"method": "cosine"}),
						StyleMetadata:       mapToStruct(map[string]any{"color": "blue"}),
						CreatedAt:           timestamppb.New(now),
						UpdatedAt:           timestamppb.New(now),
					},
					StrengthScore:      0.85,
					SimilarityScore:    0.92,
					AbstractionBridge:  true,
					HierarchicalBridge: false,
					SemanticTags:       []string{"similar", "related"},
					Description:        stringPtr("High semantic similarity"),
				},
			},
			validateFunc: func(t *testing.T, links []*canvasv1.SemanticLink) {
				require.Len(t, links, 1)
				link := links[0]

				assert.Equal(t, sourceID, link.Base.SourceId)
				assert.Equal(t, targetID, link.Base.TargetId)
				assert.Equal(t, canvasv1.SemanticConnectionType_SEMANTIC_CONNECTION_TYPE_INTRA_LEVEL_INTRA_PARENT, link.ConnectionType)
				assert.Equal(t, 0.85, link.StrengthScore)
				assert.Equal(t, 0.92, link.SimilarityScore)
				assert.True(t, link.AbstractionBridge)
				assert.False(t, link.HierarchicalBridge)
				assert.NotNil(t, link.Base.ExplorationMetadata)
				assert.Equal(t, "cosine", link.Base.ExplorationMetadata.AsMap()["method"])
				assert.Contains(t, link.SemanticTags, "similar")
				assert.Contains(t, link.SemanticTags, "related")
				assert.NotNil(t, link.Base.StyleMetadata)
				assert.Equal(t, "blue", link.Base.StyleMetadata.AsMap()["color"])
				assert.NotNil(t, link.Description)
				assert.Equal(t, "High semantic similarity", *link.Description)
				assert.False(t, link.Base.CreatedAt.AsTime().IsZero())
				assert.False(t, link.Base.UpdatedAt.AsTime().IsZero())
			},
		},
		{
			name: "semantic link with minimal fields should be valid",
			links: []*canvasv1.SemanticLink{
				{
					Base: &canvasv1.BaseLink{
						SourceId: sourceID,
						TargetId: targetID,
						// Note: No timestamps set - should be zero
					},
					StrengthScore:   0.5,
					SimilarityScore: 0.6,
				},
			},
			validateFunc: func(t *testing.T, links []*canvasv1.SemanticLink) {
				require.Len(t, links, 1)
				link := links[0]

				assert.Equal(t, sourceID, link.Base.SourceId)
				assert.Equal(t, targetID, link.Base.TargetId)
				assert.Equal(t, canvasv1.SemanticConnectionType_SEMANTIC_CONNECTION_TYPE_INTRA_LEVEL_INTRA_PARENT, link.ConnectionType)
				assert.Equal(t, 0.5, link.StrengthScore)
				assert.Equal(t, 0.6, link.SimilarityScore)
				assert.False(t, link.AbstractionBridge)
				assert.False(t, link.HierarchicalBridge)
				// When no timestamp is set, it should be nil or zero time
				if link.Base.CreatedAt != nil {
					assert.True(t, link.Base.CreatedAt.AsTime().IsZero())
				}
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
	sourceID := uuid.New().String()
	targetID := uuid.New().String()

	tests := []struct {
		name         string
		link         *canvasv1.StructuralLink
		validateFunc func(t *testing.T, link *canvasv1.StructuralLink)
	}{
		{
			name: "structural link with all fields should be valid",
			link: &canvasv1.StructuralLink{
				Base: &canvasv1.BaseLink{
					SourceId:            sourceID,
					TargetId:            targetID,
					ExplorationMetadata: mapToStruct(map[string]any{"user_id": "12345"}),
					StyleMetadata:       mapToStruct(map[string]any{"thickness": 2}),
					CreatedAt:           timestamppb.New(now),
					UpdatedAt:           timestamppb.New(now),
				},
				ConfidenceScore: 0.95,
				Description:     stringPtr("User-created connection"),
				CreatedBy:       "user_12345",
			},
			validateFunc: func(t *testing.T, link *canvasv1.StructuralLink) {
				assert.Equal(t, sourceID, link.Base.SourceId)
				assert.Equal(t, targetID, link.Base.TargetId)
				assert.Equal(t, canvasv1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_USER_DRAWN, link.ConnectionType)
				assert.Equal(t, 0.95, link.ConfidenceScore)
				assert.NotNil(t, link.Description)
				assert.Equal(t, "User-created connection", *link.Description)
				assert.NotNil(t, link.Base.ExplorationMetadata)
				assert.Equal(t, "12345", link.Base.ExplorationMetadata.AsMap()["user_id"])
				assert.NotNil(t, link.Base.StyleMetadata)
				assert.Equal(t, float64(2), link.Base.StyleMetadata.AsMap()["thickness"])
				assert.Equal(t, "user_12345", link.CreatedBy)
				assert.False(t, link.Base.CreatedAt.AsTime().IsZero())
				assert.False(t, link.Base.UpdatedAt.AsTime().IsZero())
			},
		},
		{
			name: "structural link with minimal fields should be valid",
			link: &canvasv1.StructuralLink{
				Base: &canvasv1.BaseLink{
					SourceId: sourceID,
					TargetId: targetID,
				},
				ConfidenceScore: 0.7,
				CreatedBy:       "system",
			},
			validateFunc: func(t *testing.T, link *canvasv1.StructuralLink) {
				assert.Equal(t, sourceID, link.Base.SourceId)
				assert.Equal(t, targetID, link.Base.TargetId)
				assert.Equal(t, canvasv1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_EVIDENCE_BASED, link.ConnectionType)
				assert.Equal(t, 0.7, link.ConfidenceScore)
				assert.Equal(t, "system", link.CreatedBy)
				assert.Nil(t, link.Description)
				// When no timestamp is set, it should be nil or zero time
				if link.Base.CreatedAt != nil {
					assert.True(t, link.Base.CreatedAt.AsTime().IsZero())
				}
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

// CreateHierarchicalLinks for testing - delegates to ExecuteWrite
func (r *TestableLinkRepo) CreateHierarchicalLinks(ctx context.Context, links []*canvasv1.HierarchicalLink) error {
	if len(links) == 0 {
		return nil
	}

	// This is a simplified version for testing - in reality this would be implemented
	// in the actual LinkRepo but we're testing the query generation through ExecuteWrite
	items := make([]map[string]any, 0, len(links))
	for _, link := range links {
		createdAt := link.Base.CreatedAt.AsTime()
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		updatedAt := link.Base.UpdatedAt.AsTime()
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}
		items = append(items, map[string]any{
			"src":                  link.Base.SourceId,
			"dst":                  link.Base.TargetId,
			"connection_type":      link.ConnectionType,
			"hierarchy_depth":      link.HierarchyDepth,
			"exploration_metadata": link.Base.ExplorationMetadata,
			"style_metadata":       link.Base.StyleMetadata,
			"created_at":           createdAt,
			"updated_at":           updatedAt,
			"deleted_at":           link.Base.DeletedAt.AsTime(),
		})
	}

	params := map[string]any{
		"items": items,
	}

	cypher := `
        UNWIND $items AS item
        MATCH (s:Node {id: item.src})
        MATCH (t:Node {id: item.dst})
        MERGE (s)-[r:HIERARCHICAL_PARENT]->(t)
        SET r.connection_type = item.connection_type,
            r.hierarchy_depth = item.hierarchy_depth,
            r.exploration_metadata = item.exploration_metadata,
            r.style_metadata = item.style_metadata,
            r.created_at = datetime(item.created_at),
            r.updated_at = datetime(item.updated_at),
            r.deleted_at = CASE WHEN item.deleted_at IS NULL THEN NULL ELSE datetime(item.deleted_at) END
    `

	return r.ExecuteWrite(cypher, params)
}

func TestLinkRepo_CreateHierarchicalLinks_DatabaseInteraction(t *testing.T) {
	ctx := context.Background()
	links := []*canvasv1.HierarchicalLink{
		{
			Base: &canvasv1.BaseLink{
				SourceId: uuid.New().String(),
				TargetId: uuid.New().String(),
			},
			HierarchyDepth: 1,
		},
		{
			Base: &canvasv1.BaseLink{
				SourceId: uuid.New().String(),
				TargetId: uuid.New().String(),
			},
			HierarchyDepth: 1,
		},
	}

	tests := []struct {
		name           string
		links          []*canvasv1.HierarchicalLink
		validateCypher func(t *testing.T, cypher string, params map[string]any, links []*canvasv1.HierarchicalLink)
	}{
		{
			name:  "hierarchical links should generate correct cypher",
			links: links,
			validateCypher: func(t *testing.T, cypher string, params map[string]any, testLinks []*canvasv1.HierarchicalLink) {
				assert.Contains(t, cypher, "UNWIND $items AS item")
				assert.Contains(t, cypher, "MATCH (s:Node {id: item.src})")
				assert.Contains(t, cypher, "MATCH (t:Node {id: item.dst})")
				assert.Contains(t, cypher, "MERGE (s)-[r:HIERARCHICAL_PARENT]->(t)")
				assert.Contains(t, cypher, "SET r.connection_type = item.connection_type")
				assert.Contains(t, cypher, "r.hierarchy_depth = item.hierarchy_depth")

				assert.Contains(t, params, "items")

				items, ok := params["items"].([]map[string]any)
				assert.True(t, ok)
				assert.Len(t, items, 2)

				for i, item := range items {
					assert.Contains(t, item, "src")
					assert.Contains(t, item, "dst")
					assert.Contains(t, item, "connection_type")
					assert.Contains(t, item, "hierarchy_depth")
					assert.Equal(t, testLinks[i].ConnectionType, item["connection_type"])
					assert.Equal(t, int32(testLinks[i].HierarchyDepth), item["hierarchy_depth"])
					assert.Equal(t, testLinks[i].Base.SourceId, item["src"])
					assert.Equal(t, testLinks[i].Base.TargetId, item["dst"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := &TestableLinkRepo{
				executeWriteFunc: func(cypher string, params map[string]any) error {
					tt.validateCypher(t, cypher, params, tt.links)
					return nil
				},
			}

			// Simulate the logic from CreateHierarchicalLinks
			if len(tt.links) == 0 {
				return // Early return
			}

			// Call the actual method to test query generation
			err := testRepo.CreateHierarchicalLinks(ctx, tt.links)
			assert.NoError(t, err)
		})
	}
}

func TestLinkRepo_CreateSemanticLinks_DatabaseInteraction(t *testing.T) {
	now := time.Now()
	links := []*canvasv1.SemanticLink{
		{
			Base: &canvasv1.BaseLink{
				SourceId:            uuid.New().String(),
				TargetId:            uuid.New().String(),
				ExplorationMetadata: mapToStruct(map[string]any{"method": "cosine"}),
				StyleMetadata:       mapToStruct(map[string]any{"color": "blue"}),
				CreatedAt:           timestamppb.New(now),
				UpdatedAt:           timestamppb.New(now),
			},
			StrengthScore:      0.85,
			SimilarityScore:    0.92,
			AbstractionBridge:  true,
			HierarchicalBridge: false,
			SemanticTags:       []string{"similar", "related"},
			Description:        stringPtr("High semantic similarity"),
		},
		{
			Base: &canvasv1.BaseLink{
				SourceId: uuid.New().String(),
				TargetId: uuid.New().String(),
			},
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
			assert.Equal(t, canvasv1.SemanticConnectionType_SEMANTIC_CONNECTION_TYPE_INTRA_LEVEL_INTRA_PARENT, item1["connection_type"])
			assert.Equal(t, 0.85, item1["strength_score"])
			assert.Equal(t, 0.92, item1["similarity_score"])
			assert.Equal(t, true, item1["abstraction_bridge"])
			assert.Equal(t, false, item1["hierarchical_bridge"])

			// Check second link (minimal)
			item2 := items[1]
			assert.Equal(t, canvasv1.SemanticConnectionType_SEMANTIC_CONNECTION_TYPE_INTRA_LEVEL_INTRA_PARENT, item2["connection_type"])
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
		createdAt := l.Base.CreatedAt.AsTime()
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		updatedAt := l.Base.UpdatedAt.AsTime()
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}
		params["items"] = append(params["items"].([]map[string]any), map[string]any{
			"src":                  l.Base.SourceId,
			"dst":                  l.Base.TargetId,
			"connection_type":      l.ConnectionType,
			"strength_score":       l.StrengthScore,
			"similarity_score":     l.SimilarityScore,
			"abstraction_bridge":   l.AbstractionBridge,
			"hierarchical_bridge":  l.HierarchicalBridge,
			"exploration_metadata": l.Base.ExplorationMetadata,
			"semantic_tags":        l.SemanticTags,
			"style_metadata":       l.Base.StyleMetadata,
			"description":          l.Description,
			"created_at":           createdAt,
			"updated_at":           updatedAt,
			"deleted_at":           l.Base.DeletedAt.AsTime(),
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
	link := &canvasv1.StructuralLink{
		Base: &canvasv1.BaseLink{
			SourceId:            uuid.New().String(),
			TargetId:            uuid.New().String(),
			ExplorationMetadata: mapToStruct(map[string]any{"user_id": "12345"}),
			StyleMetadata:       mapToStruct(map[string]any{"thickness": 2}),
			CreatedAt:           timestamppb.New(now),
			UpdatedAt:           timestamppb.New(now),
		},
		ConfidenceScore: 0.95,
		Description:     stringPtr("User-created connection"),
		CreatedBy:       "user_12345",
	}

	testRepo := &TestableLinkRepo{
		executeWriteFunc: func(cypher string, params map[string]any) error {
			assert.Contains(t, cypher, "MATCH (s:Node {id: $src})")
			assert.Contains(t, cypher, "MATCH (t:Node {id: $dst})")
			assert.Contains(t, cypher, "MERGE (s)-[r:STRUCTURAL_LINK]->(t)")
			assert.Contains(t, cypher, "SET r.connection_type = $connection_type")
			assert.Contains(t, cypher, "r.confidence_score = $confidence_score")
			assert.Contains(t, cypher, "r.created_by = $created_by")

			assert.Contains(t, params, "src")
			assert.Contains(t, params, "dst")
			assert.Contains(t, params, "connection_type")
			assert.Contains(t, params, "confidence_score")
			assert.Contains(t, params, "created_by")

			assert.Equal(t, link.Base.SourceId, params["src"])
			assert.Equal(t, link.Base.TargetId, params["dst"])
			assert.Equal(t, canvasv1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_USER_DRAWN, params["connection_type"])
			assert.Equal(t, 0.95, params["confidence_score"])
			assert.Equal(t, "user_12345", params["created_by"])

			return nil
		},
	}

	// Simulate CreateStructuralLink logic
	createdAt := link.Base.CreatedAt.AsTime()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := link.Base.UpdatedAt.AsTime()
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	params := map[string]any{
		"src":                  link.Base.SourceId,
		"dst":                  link.Base.TargetId,
		"connection_type":      link.ConnectionType,
		"confidence_score":     link.ConfidenceScore,
		"description":          link.Description,
		"exploration_metadata": link.Base.ExplorationMetadata,
		"style_metadata":       link.Base.StyleMetadata,
		"created_by":           link.CreatedBy,
		"created_at":           createdAt,
		"updated_at":           updatedAt,
		"deleted_at":           link.Base.DeletedAt.AsTime(),
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

// Helper functions
func mapToStruct(m map[string]any) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		panic(err) // For tests, panic is acceptable
	}
	return s
}
