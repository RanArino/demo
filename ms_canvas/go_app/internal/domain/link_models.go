package domain

import (
	"time"

	"github.com/google/uuid"
)

type SemanticLink struct {
	SourceID            uuid.UUID
	TargetID            uuid.UUID
	ConnectionType      string
	StrengthScore       float64
	SimilarityScore     float64
	AbstractionBridge   bool
	HierarchicalBridge  bool
	ExplorationMetadata map[string]any
	SemanticTags        []string
	StyleMetadata       map[string]any
	Description         *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

type StructuralLink struct {
	SourceID            uuid.UUID
	TargetID            uuid.UUID
	ConnectionType      string
	ConfidenceScore     float64
	Description         *string
	ExplorationMetadata map[string]any
	StyleMetadata       map[string]any
	CreatedBy           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

type LinkRepository interface {
	CreateHierarchicalLinks(contentSourceID uuid.UUID, chunkIDs []uuid.UUID, connectionType string, hierarchyDepth int, createdAt time.Time) error
	CreateSemanticLinks(links []SemanticLink) error
	CreateStructuralLink(link StructuralLink) error
	// Semantic link reads
	GetSemanticLink(sourceID, targetID uuid.UUID) (*SemanticLink, error)
	ListSemanticLinksFrom(nodeID uuid.UUID) ([]SemanticLink, error)
	ListSemanticLinksTo(nodeID uuid.UUID) ([]SemanticLink, error)
	// Semantic link updates/deletes
	UpdateSemanticLink(link SemanticLink) error
	DeleteSemanticLink(sourceID, targetID uuid.UUID) error
	DeleteSemanticLinksForNode(nodeID uuid.UUID) error
	// Structural link reads
	GetStructuralLink(sourceID, targetID uuid.UUID) (*StructuralLink, error)
	ListStructuralLinksFrom(nodeID uuid.UUID) ([]StructuralLink, error)
	ListStructuralLinksTo(nodeID uuid.UUID) ([]StructuralLink, error)
	// Structural link updates/deletes
	UpdateStructuralLink(link StructuralLink) error
	DeleteStructuralLink(sourceID, targetID uuid.UUID) error
	DeleteStructuralLinksForNode(nodeID uuid.UUID) error
	// Hierarchical link deletes (optional convenience)
	DeleteHierarchicalLinksByContentSource(contentSourceID uuid.UUID) error
	DeleteHierarchicalLinks(contentSourceID uuid.UUID, chunkIDs []uuid.UUID) error
}
