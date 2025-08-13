package service

import (
	"context"
	"fmt"
	"strings"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

type SpaceService struct {
	spaceRepo   domain.SpaceRepository
	contentRepo domain.ContentRepository
	graphRepo   domain.GraphRepository
}

func NewSpaceService(spaceRepo domain.SpaceRepository, contentRepo domain.ContentRepository, graphRepo domain.GraphRepository) *SpaceService {
	return &SpaceService{
		spaceRepo:   spaceRepo,
		contentRepo: contentRepo,
		graphRepo:   graphRepo,
	}
}

func (s *SpaceService) CreateSpace(ctx context.Context, title, description, ownerID string) (*domain.Space, error) {
	// Validate inputs
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	if strings.TrimSpace(ownerID) == "" {
		return nil, fmt.Errorf("owner_id is required")
	}

	// Parse ownerID as UUID
	ownerUUID, parseErr := uuid.Parse(strings.TrimSpace(ownerID))
	if parseErr != nil {
		return nil, fmt.Errorf("invalid owner_id format: %w", parseErr)
	}

	space := &domain.Space{
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		OwnerID:     ownerUUID,
	}

	err := s.spaceRepo.Create(ctx, space)
	if err != nil {
		return nil, fmt.Errorf("failed to create space: %w", err)
	}

	return space, nil
}

func (s *SpaceService) GetSpace(ctx context.Context, id uuid.UUID) (*domain.SpaceWithStats, error) {
	space, err := s.spaceRepo.GetWithStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get space: %w", err)
	}

	return space, nil
}

func (s *SpaceService) ListSpaces(ctx context.Context, filter domain.SpaceFilter) ([]*domain.SpaceWithStats, error) {
	spaces, err := s.spaceRepo.ListWithStats(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list spaces: %w", err)
	}

	return spaces, nil
}

func (s *SpaceService) UpdateSpace(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*domain.Space, error) {
	// Get existing space
	existing, err := s.spaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get space: %w", err)
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok && strings.TrimSpace(title) != "" {
		existing.Title = strings.TrimSpace(title)
	}
	if description, ok := updates["description"].(string); ok {
		existing.Description = strings.TrimSpace(description)
	}
	if ownerID, ok := updates["owner_id"].(string); ok && strings.TrimSpace(ownerID) != "" {
		ownerUUID, err := uuid.Parse(strings.TrimSpace(ownerID))
		if err != nil {
			return nil, fmt.Errorf("invalid owner_id format: %w", err)
		}
		existing.OwnerID = ownerUUID
	}

	err = s.spaceRepo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update space: %w", err)
	}

	return existing, nil
}

func (s *SpaceService) DeleteSpace(ctx context.Context, id uuid.UUID, hardDelete, force bool) error {
	// Check if space has content (unless force is true)
	if !force {
		contentCount, err := s.contentRepo.CountBySpace(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to count content: %w", err)
		}
		if contentCount > 0 {
			return fmt.Errorf("cannot delete space with content (use force=true to override)")
		}
	}

	err := s.spaceRepo.Delete(ctx, id, hardDelete)
	if err != nil {
		return fmt.Errorf("failed to delete space: %w", err)
	}

	return nil
}

func (s *SpaceService) SearchSpaces(ctx context.Context, query string, filter domain.SpaceFilter) ([]*domain.SpaceWithStats, error) {
	if strings.TrimSpace(query) == "" {
		return s.ListSpaces(ctx, filter)
	}

	spaces, err := s.spaceRepo.Search(ctx, query, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search spaces: %w", err)
	}

	// Convert to SpaceWithStats
	result := make([]*domain.SpaceWithStats, len(spaces))
	for i, space := range spaces {
		// Get stats for each space
		contentCount, err := s.contentRepo.CountBySpace(ctx, space.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get content count: %w", err)
		}

		stats := &domain.SpaceStats{
			ContentCount:   contentCount,
			LinkCount:      0,               // Will be implemented with Neo4j integration
			LastActivityAt: space.CreatedAt, // Default to creation time
		}

		result[i] = &domain.SpaceWithStats{
			Space: *space,
			Stats: *stats,
		}
	}

	return result, nil
}
