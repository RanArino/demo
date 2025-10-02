package neo4j

import (
	"strings"
	"testing"
	"time"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newStruct(m map[string]any) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		panic(err)
	}
	return s
}

func applyHierarchicalDefaults(links []*canvasv1.HierarchicalLink) []*canvasv1.HierarchicalLink {
	processed := make([]*canvasv1.HierarchicalLink, len(links))
	for i, link := range links {
		if link == nil {
			continue
		}
		// Use proto.Clone to avoid copying internal mutex/MessageState
		cloneMsg := proto.Clone(link).(*canvasv1.HierarchicalLink)
		if cloneMsg.ConnectionType == canvasv1.HierarchicalConnectionType_HIERARCHICAL_CONNECTION_TYPE_UNSPECIFIED {
			cloneMsg.ConnectionType = canvasv1.HierarchicalConnectionType_HIERARCHICAL_CONNECTION_TYPE_ABSTRACTION
		}
		processed[i] = cloneMsg
	}
	return processed
}

func applySemanticDefaults(links []*canvasv1.SemanticLink) []*canvasv1.SemanticLink {
	processed := make([]*canvasv1.SemanticLink, len(links))
	for i, link := range links {
		if link == nil {
			continue
		}
		cloneMsg := proto.Clone(link).(*canvasv1.SemanticLink)
		if cloneMsg.ConnectionType == canvasv1.SemanticConnectionType_SEMANTIC_CONNECTION_TYPE_UNSPECIFIED {
			cloneMsg.ConnectionType = canvasv1.SemanticConnectionType_SEMANTIC_CONNECTION_TYPE_INTRA_LEVEL_INTRA_PARENT
		}
		processed[i] = cloneMsg
	}
	return processed
}

func applyStructuralDefaults(link *canvasv1.StructuralLink) *canvasv1.StructuralLink {
	if link == nil {
		return nil
	}
	cloneMsg := proto.Clone(link).(*canvasv1.StructuralLink)
	if cloneMsg.ConnectionType == canvasv1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_UNSPECIFIED {
		if strings.HasPrefix(cloneMsg.CreatedBy, "user_") {
			cloneMsg.ConnectionType = canvasv1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_USER_DRAWN
		} else {
			cloneMsg.ConnectionType = canvasv1.StructuralConnectionType_STRUCTURAL_CONNECTION_TYPE_EVIDENCE_BASED
		}
	}
	return cloneMsg
}

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
				{Base: &canvasv1.BaseLink{SourceId: uuid.New().String(), TargetId: uuid.New().String()}, HierarchyDepth: 1},
				{Base: &canvasv1.BaseLink{SourceId: uuid.New().String(), TargetId: uuid.New().String()}, HierarchyDepth: 1},
				{Base: &canvasv1.BaseLink{SourceId: uuid.New().String(), TargetId: uuid.New().String()}, HierarchyDepth: 1},
			},
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
			processed := applyHierarchicalDefaults(tt.links)
			if tt.expectEmptyReturn {
				assert.Len(t, processed, 0)
			} else {
				tt.validateFunc(t, processed)
			}
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
						ExplorationMetadata: newStruct(map[string]any{"method": "cosine"}),
						StyleMetadata:       newStruct(map[string]any{"color": "blue"}),
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
			links: []*canvasv1.SemanticLink{{
				Base: &canvasv1.BaseLink{
					SourceId: sourceID,
					TargetId: targetID,
				},
				StrengthScore:   0.5,
				SimilarityScore: 0.6,
			}},
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
				if link.Base.CreatedAt != nil {
					assert.True(t, link.Base.CreatedAt.AsTime().IsZero())
				}
				assert.Nil(t, link.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processed := applySemanticDefaults(tt.links)
			tt.validateFunc(t, processed)
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
					ExplorationMetadata: newStruct(map[string]any{"user_id": "12345"}),
					StyleMetadata:       newStruct(map[string]any{"thickness": 2}),
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
				if link.Base.CreatedAt != nil {
					assert.True(t, link.Base.CreatedAt.AsTime().IsZero())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processed := applyStructuralDefaults(tt.link)
			tt.validateFunc(t, processed)
		})
	}
}
