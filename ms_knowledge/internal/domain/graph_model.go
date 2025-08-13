package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RelationType represents the type of relationship between content nodes
type RelationType string

// NOTE: Incrementally add more relation types as needed
const (
	RelationTypeReferences RelationType = "REFERENCES"
	RelationTypeContains   RelationType = "CONTAINS"
	RelationTypeRelated    RelationType = "RELATED"
	RelationTypeFollows    RelationType = "FOLLOWS"
)

// KnowledgeLink represents a relationship between two content nodes
type KnowledgeLink struct {
	ID            uuid.UUID    `json:"id"`
	FromContentID uuid.UUID    `json:"from_content_id"`
	ToContentID   uuid.UUID    `json:"to_content_id"`
	RelationType  RelationType `json:"relation_type"`
	Weight        *float64     `json:"weight,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// LinkDirection represents the direction of links to retrieve
type LinkDirection string

const (
	LinkDirectionInbound  LinkDirection = "INBOUND"
	LinkDirectionOutbound LinkDirection = "OUTBOUND"
	LinkDirectionBoth     LinkDirection = "BOTH"
)

// LinkFilter represents filters for listing knowledge links
type LinkFilter struct {
	ContentID    uuid.UUID
	Direction    LinkDirection
	RelationType RelationType
	Limit        int
	Offset       int
}

type GraphRepository interface {
	// Knowledge Link Operations (CRUD)
	CreateKnowledgeLink(ctx context.Context, link *KnowledgeLink) error
	GetKnowledgeLink(ctx context.Context, id uuid.UUID) (*KnowledgeLink, error)
	ListKnowledgeLinks(ctx context.Context, filter LinkFilter) ([]*KnowledgeLink, error)
	UpdateKnowledgeLink(ctx context.Context, link *KnowledgeLink) error
	DeleteKnowledgeLink(ctx context.Context, id uuid.UUID) error

	// Convenience Methods
	CreateLink(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID) error
	CreateLinkWithType(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID, relationshipType string) error
	GetBacklinks(ctx context.Context, contentID uuid.UUID) ([]uuid.UUID, error)

	// Content Node Lifecycle (called automatically by ContentRepository)
	// Note: Links are space-scoped, so nodes must carry spaceId
	CreateContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID) error
	DeleteContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID) error

	// Statistics
	CountLinksByContent(ctx context.Context, contentID uuid.UUID) (int64, error)
	CountLinksBySpace(ctx context.Context, spaceID uuid.UUID) (int64, error)
}
