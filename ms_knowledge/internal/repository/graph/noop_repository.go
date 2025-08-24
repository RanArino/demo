package graph

import (
	"context"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

// NoopRepository is a stub GraphRepository used when Neo4j is unavailable.
type NoopRepository struct{}

// Ensure interface implementation
var _ domain.GraphRepository = (*NoopRepository)(nil)

func NewNoopRepository() domain.GraphRepository { return &NoopRepository{} }

// Knowledge Link Operations (CRUD)
func (r *NoopRepository) CreateKnowledgeLink(ctx context.Context, link *domain.KnowledgeLink) error {
	return nil
}
func (r *NoopRepository) GetKnowledgeLink(ctx context.Context, id uuid.UUID) (*domain.EnrichedKnowledgeLink, error) {
	return nil, nil
}
func (r *NoopRepository) ListKnowledgeLinks(ctx context.Context, filter domain.LinkFilter) ([]*domain.EnrichedKnowledgeLink, error) {
	return []*domain.EnrichedKnowledgeLink{}, nil
}
func (r *NoopRepository) ListKnowledgeLinksBySpace(ctx context.Context, spaceID uuid.UUID, relationType domain.RelationType, limit, offset int) ([]*domain.KnowledgeLink, error) {
	return []*domain.KnowledgeLink{}, nil
}
func (r *NoopRepository) UpdateKnowledgeLink(ctx context.Context, link *domain.KnowledgeLink) error {
	return nil
}
func (r *NoopRepository) DeleteKnowledgeLink(ctx context.Context, id uuid.UUID) error { return nil }

// Convenience Methods
func (r *NoopRepository) CreateLink(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID) error {
	return nil
}
func (r *NoopRepository) CreateLinkWithType(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID, relationshipType string) error {
	return nil
}
func (r *NoopRepository) GetBacklinks(ctx context.Context, contentID uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{}, nil
}

// Content Node Lifecycle
func (r *NoopRepository) CreateContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID, title string, contentSummary *string) error {
	return nil
}
func (r *NoopRepository) DeleteContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID) error {
	return nil
}

// Statistics
func (r *NoopRepository) CountLinksByContent(ctx context.Context, contentID uuid.UUID) (int64, error) {
	return 0, nil
}
func (r *NoopRepository) CountLinksBySpace(ctx context.Context, spaceID uuid.UUID) (int64, error) {
	return 0, nil
}
