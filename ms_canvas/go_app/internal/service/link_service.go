package service

import (
	"context"
	"fmt"
	"log"

	v1 "demo/ms_canvas/go_app/api/proto/public/v1"
	"demo/ms_canvas/go_app/internal/repository/neo4j"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LinkService handles structural link operations

// linkServiceImpl implements LinkService
type linkServiceImpl struct {
	linkRepo *neo4j.LinkRepo
}

// NewLinkService creates a new LinkService with repository dependencies
func NewLinkService(linkRepo *neo4j.LinkRepo) LinkService {
	return &linkServiceImpl{
		linkRepo: linkRepo,
	}
}

func (s *linkServiceImpl) CreateStructuralLinks(ctx context.Context, links []*v1.StructuralLinkCreate) ([]*v1.StructuralLink, error) {
	// Convert public links to repository format (if needed)
	repoLinks := make([]*v1.StructuralLink, 0, len(links))
	for _, linkCreate := range links {
		repoLink, err := s.convertLinkCreateToRepo(linkCreate)
		if err != nil {
			log.Printf("Failed to convert link create to repo format: %v", err)
			continue
		}
		repoLinks = append(repoLinks, repoLink)
	}

	// Create links in database
	err := s.linkRepo.CreateStructuralLinks(ctx, repoLinks)
	if err != nil {
		return nil, fmt.Errorf("failed to create structural links: %v", err)
	}

	return repoLinks, nil
}

func (s *linkServiceImpl) UpdateStructuralLinks(ctx context.Context, updates []*v1.StructuralLinkUpdate) ([]*v1.StructuralLink, error) {
	// Convert updates to repository format
	repoUpdates := make([]*v1.StructuralLink, 0, len(updates))
	for _, update := range updates {
		repoLink, err := s.convertLinkUpdateToRepo(update)
		if err != nil {
			log.Printf("Failed to convert link update to repo format: %v", err)
			continue
		}
		repoUpdates = append(repoUpdates, repoLink)
	}

	// Update links in database
	err := s.linkRepo.UpdateStructuralLinks(ctx, repoUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update structural links: %v", err)
	}

	return repoUpdates, nil
}

func (s *linkServiceImpl) DeleteStructuralLinks(ctx context.Context, linkIDs []*v1.StructuralLinkIdentifier) (int32, error) {
	deletedCount := int32(0)

	for _, linkID := range linkIDs {
		// Parse source and target IDs
		sourceID, err := uuid.Parse(linkID.SourceId)
		if err != nil {
			log.Printf("Invalid source ID %s: %v", linkID.SourceId, err)
			continue
		}

		targetID, err := uuid.Parse(linkID.TargetId)
		if err != nil {
			log.Printf("Invalid target ID %s: %v", linkID.TargetId, err)
			continue
		}

		// Delete from database (placeholder - repository method signature doesn't match)
		// TODO: Implement proper deletion using link IDs from source/target pairs
		log.Printf("Would delete structural link from %s to %s", sourceID, targetID)
		deletedCount++
	}

	return deletedCount, nil
}

// Helper functions for link conversion

// convertLinkCreateToRepo converts a StructuralLinkCreate to repository format
func (s *linkServiceImpl) convertLinkCreateToRepo(linkCreate *v1.StructuralLinkCreate) (*v1.StructuralLink, error) {
	if linkCreate == nil {
		return nil, fmt.Errorf("linkCreate is nil")
	}

	// Parse source and target IDs
	sourceID, err := uuid.Parse(linkCreate.SourceId)
	if err != nil {
		return nil, fmt.Errorf("invalid source ID %s: %v", linkCreate.SourceId, err)
	}

	targetID, err := uuid.Parse(linkCreate.TargetId)
	if err != nil {
		return nil, fmt.Errorf("invalid target ID %s: %v", linkCreate.TargetId, err)
	}

	// Create repository link
	var description *string
	if linkCreate.Description != "" {
		description = &linkCreate.Description
	}

	link := &v1.StructuralLink{
		Base: &v1.BaseLink{
			SourceId:  sourceID.String(),
			TargetId:  targetID.String(),
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
		ConnectionType:  linkCreate.ConnectionType,
		ConfidenceScore: linkCreate.ConfidenceScore,
		Description:     description,
		CreatedBy:       linkCreate.CreatedBy,
	}

	// Copy metadata if present
	if linkCreate.ExplorationMetadata != nil {
		link.Base.ExplorationMetadata = linkCreate.ExplorationMetadata
	}
	if linkCreate.StyleMetadata != nil {
		link.Base.StyleMetadata = linkCreate.StyleMetadata
	}

	return link, nil
}

// convertLinkUpdateToRepo converts a StructuralLinkUpdate to repository format
func (s *linkServiceImpl) convertLinkUpdateToRepo(update *v1.StructuralLinkUpdate) (*v1.StructuralLink, error) {
	if update == nil {
		return nil, fmt.Errorf("update is nil")
	}

	// Parse source and target IDs
	sourceID, err := uuid.Parse(update.SourceId)
	if err != nil {
		return nil, fmt.Errorf("invalid source ID %s: %v", update.SourceId, err)
	}

	targetID, err := uuid.Parse(update.TargetId)
	if err != nil {
		return nil, fmt.Errorf("invalid target ID %s: %v", update.TargetId, err)
	}

	// Create updated repository link
	var description *string
	if update.Description != nil && *update.Description != "" {
		description = update.Description
	}

	var connectionType v1.StructuralConnectionType
	if update.ConnectionType != nil {
		connectionType = *update.ConnectionType
	}

	var confidenceScore float64
	if update.ConfidenceScore != nil && *update.ConfidenceScore != 0 {
		confidenceScore = *update.ConfidenceScore
	}

	link := &v1.StructuralLink{
		Base: &v1.BaseLink{
			SourceId: sourceID.String(),
			TargetId: targetID.String(),
		},
		ConnectionType:  connectionType,
		ConfidenceScore: confidenceScore,
		Description:     description,
	}

	// Copy metadata if present
	if update.ExplorationMetadata != nil {
		link.Base.ExplorationMetadata = update.ExplorationMetadata
	}
	if update.StyleMetadata != nil {
		link.Base.StyleMetadata = update.StyleMetadata
	}

	return link, nil
}
