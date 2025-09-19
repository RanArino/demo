package domain

import (
	"time"

	"github.com/google/uuid"
)

type BaseNode struct {
	ID               uuid.UUID
	SpaceID          uuid.UUID
	AbstractionLevel int
	ContextType      string
	Embedding        []float32
	Keywords         []string
	ML               *MLInfo
	ChatContent      *string
	DisplayContent   *string
	SemanticDensity  *float64
	Position3D       *SpatialCoordinates
	IsPositionLocked *bool
	Visibility       *bool
	DisplayProps     *DisplayProps
	EngagementScore  *EngagementScore
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type ChunkNode struct {
	BaseNode
	ContentSourceID uuid.UUID
	SequenceIndex   int
	ChunkType       string
	StartPosition   int64
	EndPosition     int64
	TokenCount      *int
	Content         string
}

type ContentNode struct {
	BaseNode
	ContentSourceID uuid.UUID
	Title           *string
	MediaType       *string
	Source          *string
	TokenCount      *int
	ActionData      map[string]interface{}
}

type ClusterNode struct {
	BaseNode
	ClusterScope  string
	Title         *string
	MemberCount   *int
	CoverageScore *float64
}

type SpatialCoordinates struct {
	X int
	Y int
	Z int
}

type DisplayProps struct {
	Size    int
	Opacity float64
	Shape   string
	Color   string
}

type EngagementScore struct {
	CanvasScore  float64
	ChatScore    float64
	OverallScore float64
}

type MLInfo struct {
	ClusteringModelVersionID *uuid.UUID
	DRModelVersionID         *uuid.UUID
	EmbeddingModelID         *string
	ModelVersion             *string
}

type MLInfoUpdate struct {
	ClusteringModelVersionID *uuid.UUID `json:"clustering_model_version_id,omitempty"`
	DRModelVersionID         *uuid.UUID `json:"dr_model_version_id,omitempty"`
	EmbeddingModelID         *string    `json:"embedding_model_id,omitempty"`
	ModelVersion             *string    `json:"model_version,omitempty"`
}

type BaseNodeUpdate struct {
	Embedding        *[]float32          `json:"embedding,omitempty"`
	Keywords         *[]string           `json:"keywords,omitempty"`
	ML               *MLInfoUpdate       `json:"ml,omitempty"`
	ChatContent      *string             `json:"chat_content,omitempty"`
	DisplayContent   *string             `json:"display_content,omitempty"`
	SemanticDensity  *float64            `json:"semantic_density,omitempty"`
	Position3D       *SpatialCoordinates `json:"position_3d,omitempty"`
	IsPositionLocked *bool               `json:"is_position_locked,omitempty"`
	Visibility       *bool               `json:"visibility,omitempty"`
	DisplayProps     *DisplayProps       `json:"display_props,omitempty"`
	EngagementScore  *EngagementScore    `json:"engagement_score,omitempty"`
}

type ClusterNodeUpdate struct {
	ID uuid.UUID `json:"id"`
	BaseNodeUpdate
	ClusterScope  *string  `json:"cluster_scope,omitempty"`
	Title         *string  `json:"title,omitempty"`
	MemberCount   *int     `json:"member_count,omitempty"`
	CoverageScore *float64 `json:"coverage_score,omitempty"`
}

type ContentNodeUpdate struct {
	ID uuid.UUID `json:"id"`
	BaseNodeUpdate
	Title      *string                 `json:"title,omitempty"`
	MediaType  *string                 `json:"media_type,omitempty"`
	Source     *string                 `json:"source,omitempty"`
	TokenCount *int                    `json:"token_count,omitempty"`
	ActionData *map[string]interface{} `json:"action_data,omitempty"`
}

type ChunkNodeUpdate struct {
	ID uuid.UUID `json:"id"`
	BaseNodeUpdate
	Content       *string `json:"content,omitempty"`
	SequenceIndex *int    `json:"sequence_index,omitempty"`
	ChunkType     *string `json:"chunk_type,omitempty"`
	StartPosition *int64  `json:"start_position,omitempty"`
	EndPosition   *int64  `json:"end_position,omitempty"`
	TokenCount    *int    `json:"token_count,omitempty"`
}

func (u *BaseNodeUpdate) ApplyToBase(target *BaseNode) {
	if u == nil || target == nil {
		return
	}
	if u.Embedding != nil {
		target.Embedding = *u.Embedding
	}
	if u.Keywords != nil {
		target.Keywords = *u.Keywords
	}
	if u.ChatContent != nil {
		target.ChatContent = u.ChatContent
	}
	if u.DisplayContent != nil {
		target.DisplayContent = u.DisplayContent
	}
	if u.SemanticDensity != nil {
		target.SemanticDensity = u.SemanticDensity
	}
	if u.Position3D != nil {
		target.Position3D = u.Position3D
	}
	if u.IsPositionLocked != nil {
		target.IsPositionLocked = u.IsPositionLocked
	}
	if u.Visibility != nil {
		target.Visibility = u.Visibility
	}
	if u.DisplayProps != nil {
		target.DisplayProps = u.DisplayProps
	}
	if u.EngagementScore != nil {
		target.EngagementScore = u.EngagementScore
	}
	if u.ML != nil {
		if target.ML == nil {
			target.ML = &MLInfo{}
		}
		if u.ML.ClusteringModelVersionID != nil {
			target.ML.ClusteringModelVersionID = u.ML.ClusteringModelVersionID
		}
		if u.ML.DRModelVersionID != nil {
			target.ML.DRModelVersionID = u.ML.DRModelVersionID
		}
		if u.ML.EmbeddingModelID != nil {
			target.ML.EmbeddingModelID = u.ML.EmbeddingModelID
		}
		if u.ML.ModelVersion != nil {
			target.ML.ModelVersion = u.ML.ModelVersion
		}
	}
}

func (u *ClusterNodeUpdate) Apply(target *ClusterNode) {
	if u == nil || target == nil {
		return
	}
	u.BaseNodeUpdate.ApplyToBase(&target.BaseNode)
	if u.ClusterScope != nil {
		target.ClusterScope = *u.ClusterScope
	}
	if u.Title != nil {
		target.Title = u.Title
	}
	if u.MemberCount != nil {
		target.MemberCount = u.MemberCount
	}
	if u.CoverageScore != nil {
		target.CoverageScore = u.CoverageScore
	}
}

func (u *ContentNodeUpdate) Apply(target *ContentNode) {
	if u == nil || target == nil {
		return
	}
	u.BaseNodeUpdate.ApplyToBase(&target.BaseNode)
	if u.Title != nil {
		target.Title = u.Title
	}
	if u.MediaType != nil {
		target.MediaType = u.MediaType
	}
	if u.Source != nil {
		target.Source = u.Source
	}
	if u.TokenCount != nil {
		target.TokenCount = u.TokenCount
	}
	if u.ActionData != nil {
		target.ActionData = *u.ActionData
	}
}

func (u *ChunkNodeUpdate) Apply(target *ChunkNode) {
	if u == nil || target == nil {
		return
	}
	u.BaseNodeUpdate.ApplyToBase(&target.BaseNode)
	if u.Content != nil {
		target.Content = *u.Content
	}
	if u.SequenceIndex != nil {
		target.SequenceIndex = *u.SequenceIndex
	}
	if u.ChunkType != nil {
		target.ChunkType = *u.ChunkType
	}
	if u.StartPosition != nil {
		target.StartPosition = *u.StartPosition
	}
	if u.EndPosition != nil {
		target.EndPosition = *u.EndPosition
	}
	if u.TokenCount != nil {
		target.TokenCount = u.TokenCount
	}
}

type NodeRepository interface {
	CreateChunkNodes(chunks []ChunkNode) error
	CreateContentNode(node ContentNode) error
	CreateClusterNode(id uuid.UUID, spaceID uuid.UUID, abstractionLevel int, clusterScope string, title *string, embedding *[]float32) error
	UpdateClusterNode(update ClusterNodeUpdate) error
	UpdateContentNode(update ContentNodeUpdate) error
	UpdateChunkNode(update ChunkNodeUpdate) error
	SoftDeleteNode(id uuid.UUID) error
}
