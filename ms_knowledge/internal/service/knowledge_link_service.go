package service

import (
	"context"
	"fmt"
	"time"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

type KnowledgeLinkService struct {
	graphRepo   domain.GraphRepository
	contentRepo domain.ContentRepository
}

func NewKnowledgeLinkService(graphRepo domain.GraphRepository, contentRepo domain.ContentRepository) *KnowledgeLinkService {
	return &KnowledgeLinkService{
		graphRepo:   graphRepo,
		contentRepo: contentRepo,
	}
}

func (s *KnowledgeLinkService) CreateKnowledgeLink(ctx context.Context, fromContentID, toContentID uuid.UUID, relationType domain.RelationType, weight float64) (*domain.KnowledgeLink, error) {
	// Validate inputs
	if fromContentID == uuid.Nil {
		return nil, fmt.Errorf("from_content_id is required")
	}
	if toContentID == uuid.Nil {
		return nil, fmt.Errorf("to_content_id is required")
	}
	if fromContentID == toContentID {
		return nil, fmt.Errorf("cannot create self-link")
	}

	// Validate relation type
	if relationType == "" {
		relationType = domain.RelationTypeReferences
	}

	// Clamp weight to [0,1]
	if weight < 0 {
		weight = 0
	}
	if weight > 1 {
		weight = 1
	}

	// Load content sources and enforce same-space constraint
	fromContent, err := s.contentRepo.GetByID(ctx, fromContentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load from content: %w", err)
	}
	toContent, err := s.contentRepo.GetByID(ctx, toContentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load to content: %w", err)
	}
	if fromContent.SpaceID != toContent.SpaceID {
		return nil, fmt.Errorf("cannot create link across spaces")
	}

	// Create knowledge link
	link := &domain.KnowledgeLink{
		ID:            uuid.New(),
		FromContentID: fromContentID,
		ToContentID:   toContentID,
		RelationType:  relationType,
		Weight:        &weight,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.graphRepo.CreateKnowledgeLink(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge link: %w", err)
	}

	return link, nil
}

func (s *KnowledgeLinkService) GetKnowledgeLink(ctx context.Context, id uuid.UUID) (*domain.EnrichedKnowledgeLink, error) {
	link, err := s.graphRepo.GetKnowledgeLink(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge link: %w", err)
	}
	return link, nil
}

func (s *KnowledgeLinkService) ListKnowledgeLinks(ctx context.Context, filter domain.LinkFilter) ([]*domain.EnrichedKnowledgeLink, error) {
	links, err := s.graphRepo.ListKnowledgeLinks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list knowledge links: %w", err)
	}
	return links, nil
}

func (s *KnowledgeLinkService) UpdateKnowledgeLink(ctx context.Context, id uuid.UUID, relationType domain.RelationType, weight float64) (*domain.KnowledgeLink, error) {
	// Get existing link (only for updating metadata stored on relationship)
	basic, err := s.graphRepo.GetKnowledgeLink(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge link: %w", err)
	}
	// We need a basic struct to pass to repo update; reconstruct minimal
	link := &domain.KnowledgeLink{ID: id, RelationType: relationType, Weight: &weight, CreatedAt: basic.CreatedAt, UpdatedAt: time.Now()}
	if relationType == "" {
		link.RelationType = basic.RelationType
	}
	if weight < 0 || weight > 1 {
		link.Weight = basic.Weight
	}

	err = s.graphRepo.UpdateKnowledgeLink(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("failed to update knowledge link: %w", err)
	}
	return link, nil
}

func (s *KnowledgeLinkService) DeleteKnowledgeLink(ctx context.Context, id uuid.UUID) error {
	err := s.graphRepo.DeleteKnowledgeLink(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge link: %w", err)
	}
	return nil
}

func (s *KnowledgeLinkService) GetBacklinks(ctx context.Context, contentID uuid.UUID) ([]*domain.KnowledgeLink, error) {
	ids, err := s.graphRepo.GetBacklinks(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backlinks ids: %w", err)
	}
	var res []*domain.KnowledgeLink
	for _, id := range ids {
		res = append(res, &domain.KnowledgeLink{FromContentID: id, ToContentID: contentID})
	}
	return res, nil
}

func (s *KnowledgeLinkService) CountLinksByContent(ctx context.Context, contentID uuid.UUID) (int64, error) {
	count, err := s.graphRepo.CountLinksByContent(ctx, contentID)
	if err != nil {
		return 0, fmt.Errorf("failed to count links: %w", err)
	}
	return count, nil
}

// Bulk list by space (simple)
func (s *KnowledgeLinkService) ListKnowledgeLinksBySpace(ctx context.Context, spaceID uuid.UUID, relationType domain.RelationType, limit, offset int) ([]*domain.KnowledgeLink, error) {
	links, err := s.graphRepo.ListKnowledgeLinksBySpace(ctx, spaceID, relationType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list links by space: %w", err)
	}
	return links, nil
}
