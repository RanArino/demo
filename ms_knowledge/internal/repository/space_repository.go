package repository

import (
	"context"
	"time"

	"demo/ms_knowledge/ent"
	"demo/ms_knowledge/ent/contentsource"
	"demo/ms_knowledge/ent/space"
	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

type spaceRepository struct {
	client    *ent.Client
	graphRepo domain.GraphRepository
}

// NewSpaceRepository creates a new space repository
func NewSpaceRepository(client *ent.Client, graphRepo domain.GraphRepository) domain.SpaceRepository {
	return &spaceRepository{
		client:    client,
		graphRepo: graphRepo,
	}
}

func (r *spaceRepository) Create(ctx context.Context, space *domain.Space) error {
	space.ID = uuid.New()
	space.CreatedAt = time.Now()
	space.LastUpdatedAt = time.Now()

	create := r.client.Space.Create().
		SetID(space.ID).
		SetTitle(space.Title).
		SetDescription(space.Description).
		SetIcon(space.Icon).
		SetCoverImage(space.CoverImage).
		SetKeywords(space.Keywords).
		SetOwnerID(space.OwnerID).
		SetCreatedAt(space.CreatedAt).
		SetCreatedBy(space.CreatedBy).
		SetLastUpdatedAt(space.LastUpdatedAt).
		SetAccessLevel(space.AccessLevel).
		SetGuestAccessEnabled(space.GuestAccessEnabled).
		SetStatus(space.Status).
		SetProcessingStatus(space.ProcessingStatus)

	var lastUpdatedByPtr *uuid.UUID
	if space.LastUpdatedBy != uuid.Nil {
		lastUpdatedByPtr = &space.LastUpdatedBy
	}
	create = create.SetNillableLastUpdatedBy(lastUpdatedByPtr)
	create = create.SetNillableGuestAccessExpiry(space.GuestAccessExpiry)
	create = create.SetNillableDeletedAt(space.DeletedAt)

	_, err := create.Save(ctx)
	return err
}

func (r *spaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Space, error) {
	space, err := r.client.Space.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	domainSpace := &domain.Space{
		ID:                 space.ID,
		Title:              space.Title,
		Description:        space.Description,
		Icon:               space.Icon,
		CoverImage:         space.CoverImage,
		Keywords:           space.Keywords,
		OwnerID:            space.OwnerID,
		CreatedAt:          space.CreatedAt,
		CreatedBy:          space.CreatedBy,
		LastUpdatedAt:      space.LastUpdatedAt,
		AccessLevel:        space.AccessLevel,
		GuestAccessEnabled: space.GuestAccessEnabled,
		Status:             space.Status,
		ProcessingStatus:   space.ProcessingStatus,
	}

	domainSpace.LastUpdatedBy = space.LastUpdatedBy
	if space.GuestAccessExpiry != nil {
		domainSpace.GuestAccessExpiry = space.GuestAccessExpiry
	}
	if space.DeletedAt != nil {
		domainSpace.DeletedAt = space.DeletedAt
	}

	return domainSpace, nil
}

func (r *spaceRepository) GetWithStats(ctx context.Context, id uuid.UUID) (*domain.SpaceWithStats, error) {
	space, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get content count and size statistics
	contentCount, err := r.client.ContentSource.Query().
		Where(contentsource.SpaceID(id)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	// Get total size of all content in this space
	contentSources, err := r.client.ContentSource.Query().
		Where(contentsource.SpaceID(id)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var totalSizeBytes int64
	contentByStatus := make(map[string]int64)
	processingStats := domain.ContentProcessingStats{}

	for _, source := range contentSources {
		totalSizeBytes += source.SizeBytes
		contentByStatus[source.Status]++
		
		// Update processing stats
		switch source.Status {
		case string(domain.ContentStatusUploading):
			processingStats.UploadingCount++
		case string(domain.ContentStatusUploaded):
			processingStats.UploadedCount++
		case string(domain.ContentStatusProcessing):
			processingStats.ProcessingCount++
		case string(domain.ContentStatusProcessed):
			processingStats.ProcessedCount++
		case string(domain.ContentStatusFailed):
			processingStats.FailedCount++
		case string(domain.ContentStatusPending):
			processingStats.PendingCount++
		}
	}

	// NOTE: Graph repo is temporarily disabled
	// // Get link count from Neo4j using the graph repository
	var linkCount int64
	// if r.graphRepo != nil {
	// 	linkCount, err = r.graphRepo.CountLinksBySpace(ctx, id)
	// 	if err != nil {
	// 		// Log the error but don't fail the entire operation
	// 		// This allows the space to be retrieved even if graph operations fail
	// 		linkCount = 0
	// 	}
	// } else {
	// 	linkCount = 0
	// }

	// Get last activity (most recent content source update)
	lastContent, err := r.client.ContentSource.Query().
		Where(contentsource.SpaceID(id)).
		Order(ent.Desc(contentsource.FieldUpdatedAt)).
		First(ctx)

	var lastActivityAt time.Time
	if err == nil {
		lastActivityAt = lastContent.UpdatedAt
	} else {
		lastActivityAt = space.CreatedAt
	}

	stats := domain.SpaceStats{
		ContentCount:      int64(contentCount),
		LinkCount:         linkCount,
		TotalSizeBytes:    totalSizeBytes,
		LastActivityAt:    lastActivityAt,
		ContentByStatus:   contentByStatus,
		ProcessingStats:   processingStats,
	}

	return &domain.SpaceWithStats{
		Space: *space,
		Stats: stats,
	}, nil
}

func (r *spaceRepository) List(ctx context.Context, filter domain.SpaceFilter) ([]*domain.Space, error) {
	query := r.client.Space.Query()

	if filter.OwnerID != uuid.Nil {
		query = query.Where(space.OwnerID(filter.OwnerID))
	}

	// Handle keywords filter
	if len(filter.Keywords) > 0 {
		for _, keyword := range filter.Keywords {
			query = query.Where(
				space.Or(
					space.TitleContains(keyword),
					space.DescriptionContains(keyword),
				),
			)
		}
	}

	// Handle general query (searches title, description, and keywords)
	if filter.Query != "" {
		query = query.Where(
			space.Or(
				space.TitleContains(filter.Query),
				space.DescriptionContains(filter.Query),
			),
		)
	}

	// Handle access level filter
	if filter.AccessLevel != "" {
		query = query.Where(space.AccessLevel(filter.AccessLevel))
	}

	// Handle status filter
	if filter.Status != "" {
		query = query.Where(space.Status(filter.Status))
	}

	// Handle datetime filters
	if filter.CreatedAfter != nil {
		query = query.Where(space.CreatedAtGTE(*filter.CreatedAfter))
	}
	if filter.CreatedBefore != nil {
		query = query.Where(space.CreatedAtLTE(*filter.CreatedBefore))
	}
	if filter.UpdatedAfter != nil {
		query = query.Where(space.LastUpdatedAtGTE(*filter.UpdatedAfter))
	}
	if filter.UpdatedBefore != nil {
		query = query.Where(space.LastUpdatedAtLTE(*filter.UpdatedBefore))
	}

	query = query.Order(ent.Desc(space.FieldCreatedAt))

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	spaces, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Space, len(spaces))
	for i, s := range spaces {
		domainSpace := &domain.Space{
			ID:                 s.ID,
			Title:              s.Title,
			Description:        s.Description,
			Icon:               s.Icon,
			CoverImage:         s.CoverImage,
			Keywords:           s.Keywords,
			OwnerID:            s.OwnerID,
			CreatedAt:          s.CreatedAt,
			CreatedBy:          s.CreatedBy,
			LastUpdatedAt:      s.LastUpdatedAt,
			LastUpdatedBy:      s.LastUpdatedBy,
			AccessLevel:        s.AccessLevel,
			GuestAccessEnabled: s.GuestAccessEnabled,
			Status:             s.Status,
			ProcessingStatus:   s.ProcessingStatus,
		}

		if s.GuestAccessExpiry != nil {
			domainSpace.GuestAccessExpiry = s.GuestAccessExpiry
		}
		if s.DeletedAt != nil {
			domainSpace.DeletedAt = s.DeletedAt
		}

		result[i] = domainSpace
	}

	return result, nil
}

func (r *spaceRepository) ListWithStats(ctx context.Context, filter domain.SpaceFilter) ([]*domain.SpaceWithStats, error) {
	spaces, err := r.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.SpaceWithStats, len(spaces))
	for i, s := range spaces {
		stats, err := r.getSpaceStats(ctx, s.ID)
		if err != nil {
			return nil, err
		}

		result[i] = &domain.SpaceWithStats{
			Space: *s,
			Stats: *stats,
		}
	}

	return result, nil
}

func (r *spaceRepository) Update(ctx context.Context, space *domain.Space) error {
	space.LastUpdatedAt = time.Now()

	update := r.client.Space.UpdateOneID(space.ID)

	if space.Title != "" {
		update = update.SetTitle(space.Title)
	}

	if space.Description != "" {
		update = update.SetDescription(space.Description)
	}

	if space.Icon != "" {
		update = update.SetIcon(space.Icon)
	}

	if space.CoverImage != "" {
		update = update.SetCoverImage(space.CoverImage)
	}

	if space.Keywords != nil {
		update = update.SetKeywords(space.Keywords)
	}

	if space.OwnerID != uuid.Nil {
		update = update.SetOwnerID(space.OwnerID)
	}

	if space.LastUpdatedBy != uuid.Nil {
		update = update.SetLastUpdatedBy(space.LastUpdatedBy)
	}

	update = update.SetLastUpdatedAt(space.LastUpdatedAt)

	_, err := update.Save(ctx)
	return err
}

func (r *spaceRepository) Delete(ctx context.Context, id uuid.UUID, hardDelete bool) error {
	if hardDelete {
		return r.client.Space.DeleteOneID(id).Exec(ctx)
	}

	// Soft delete - update deleted_at field (assuming it exists in the schema)
	// For now, we'll do a hard delete
	return r.client.Space.DeleteOneID(id).Exec(ctx)
}

func (r *spaceRepository) Search(ctx context.Context, query string, filter domain.SpaceFilter) ([]*domain.Space, error) {
	// Use the same logic as List but with the search query
	filter.Query = query
	return r.List(ctx, filter)
}

func (r *spaceRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.client.Space.Query().Where(space.ID(id)).Exist(ctx)
}

func (r *spaceRepository) getSpaceStats(ctx context.Context, spaceID uuid.UUID) (*domain.SpaceStats, error) {
	// Get content count and total size
	contentCount, err := r.client.ContentSource.Query().
		Where(contentsource.SpaceID(spaceID)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate total size bytes (this would need to be implemented with blob registry)
	// For now, we'll set it to 0 and it should be calculated from KNOWLEDGE_CONTENT_BLOBS
	totalSizeBytes := int64(0)

	// Get last activity
	lastContent, err := r.client.ContentSource.Query().
		Where(contentsource.SpaceID(spaceID)).
		Order(ent.Desc(contentsource.FieldUpdatedAt)).
		First(ctx)

	var lastActivityAt time.Time
	if err == nil {
		lastActivityAt = lastContent.UpdatedAt
	} else {
		// If no content, use space creation time
		space, err := r.GetByID(ctx, spaceID)
		if err != nil {
			return nil, err
		}
		lastActivityAt = space.CreatedAt
	}

	// NOTE: Graph repo is temporarily disabled
	// // Get link count from Neo4j using the graph repository
	var linkCount int64
	// if r.graphRepo != nil {
	// 	linkCount, err = r.graphRepo.CountLinksBySpace(ctx, spaceID)
	// 	if err != nil {
	// 		// Log the error but don't fail the entire operation
	// 		linkCount = 0
	// 	}
	// } else {
	// 	linkCount = 0
	// }

	return &domain.SpaceStats{
		ContentCount:   int64(contentCount),
		LinkCount:      linkCount,
		TotalSizeBytes: totalSizeBytes,
		LastActivityAt: lastActivityAt,
	}, nil
}
