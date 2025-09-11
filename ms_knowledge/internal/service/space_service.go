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
}

func NewSpaceService(spaceRepo domain.SpaceRepository, contentRepo domain.ContentRepository) *SpaceService {
	return &SpaceService{
		spaceRepo:   spaceRepo,
		contentRepo: contentRepo,
	}
}

// enforceRLSForSpace ensures the caller owns the given space owner UUID.
func (s *SpaceService) enforceRLSForSpace(ctx context.Context, spaceOwner uuid.UUID) error {
	return EnforceOwner(ctx, spaceOwner)
}

// scopeSpaceFilterToOwner forces the filter to the caller's owner_id.
func (s *SpaceService) scopeSpaceFilterToOwner(ctx context.Context, filter domain.SpaceFilter) (domain.SpaceFilter, error) {
	ownerUUID, err := MustGetOwnerUUID(ctx)
	if err != nil {
		return filter, err
	}
	filter.OwnerID = ownerUUID
	return filter, nil
}

func (s *SpaceService) CreateSpace(ctx context.Context, title, description string, keywords []string, icon string) (*domain.Space, error) {
	// Validate inputs
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	// Resolve owner ID from context (set by gRPC layer)
	ownerUUID, parseErr := MustGetOwnerUUID(ctx)
	if parseErr != nil {
		return nil, parseErr
	}

	space := &domain.Space{
		Title:         strings.TrimSpace(title),
		Description:   strings.TrimSpace(description),
		Keywords:      keywords,
		Icon:          strings.TrimSpace(icon),
		OwnerID:       ownerUUID,
		CreatedBy:     ownerUUID,
		LastUpdatedBy: ownerUUID,
		AccessLevel:   "private",
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

	// Enforce RLS: caller must be the owner
	if err := s.enforceRLSForSpace(ctx, space.Space.OwnerID); err != nil {
		return nil, err
	}

	return space, nil
}

func (s *SpaceService) ListSpaces(ctx context.Context, filter domain.SpaceFilter) ([]*domain.SpaceWithStats, error) {
	var err error
	filter, err = s.scopeSpaceFilterToOwner(ctx, filter)
	if err != nil {
		return nil, err
	}

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

	// Enforce RLS: caller must be the owner
	if err := s.enforceRLSForSpace(ctx, existing.OwnerID); err != nil {
		return nil, err
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok && strings.TrimSpace(title) != "" {
		existing.Title = strings.TrimSpace(title)
	}
	if description, ok := updates["description"].(string); ok {
		existing.Description = strings.TrimSpace(description)
	}
	if keywords, ok := updates["keywords"].([]string); ok {
		existing.Keywords = keywords
	}
	if icon, ok := updates["icon"].(string); ok && strings.TrimSpace(icon) != "" {
		existing.Icon = strings.TrimSpace(icon)
	}
	if accessLevel, ok := updates["access_level"].(string); ok && strings.TrimSpace(accessLevel) != "" {
		existing.AccessLevel = strings.TrimSpace(accessLevel)
	}
	// Disallow changing owner_id via update

	err = s.spaceRepo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update space: %w", err)
	}

	return existing, nil
}

func (s *SpaceService) DeleteSpace(ctx context.Context, id uuid.UUID, hardDelete, force bool) error {
	// Ensure the space belongs to caller before proceeding
	existing, err := s.spaceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get space: %w", err)
	}
	if err := s.enforceRLSForSpace(ctx, existing.OwnerID); err != nil {
		return err
	}

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

	err = s.spaceRepo.Delete(ctx, id, hardDelete)
	if err != nil {
		return fmt.Errorf("failed to delete space: %w", err)
	}

	return nil
}

func (s *SpaceService) SearchSpaces(ctx context.Context, query string, filter domain.SpaceFilter) ([]*domain.SpaceWithStats, error) {
	if strings.TrimSpace(query) == "" {
		return s.ListSpaces(ctx, filter)
	}
	// Scope strictly to caller
	var err error
	filter, err = s.scopeSpaceFilterToOwner(ctx, filter)
	if err != nil {
		return nil, err
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
			LinkCount:      0,               // Graph features disabled
			LastActivityAt: space.CreatedAt, // Default to creation time
		}

		result[i] = &domain.SpaceWithStats{
			Space: *space,
			Stats: *stats,
		}
	}

	return result, nil
}
