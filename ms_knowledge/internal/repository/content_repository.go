package repository

import (
	"context"
	"time"

	"demo/ms_knowledge/ent"
	"demo/ms_knowledge/ent/contentsource"
	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

type contentRepository struct {
	client *ent.Client
}

// NewContentRepository creates a new content repository
func NewContentRepository(client *ent.Client) domain.ContentRepository {
	return &contentRepository{
		client: client,
	}
}

func (r *contentRepository) Create(ctx context.Context, content *domain.ContentSource) error {
	content.ID = uuid.New()
	content.CreatedAt = time.Now()
	content.UpdatedAt = time.Now()

	create := r.client.ContentSource.Create().
		SetID(content.ID).
		SetSpaceID(content.SpaceID).
		SetOwnerID(content.OwnerID).
		SetTitle(content.Title).
		SetMediaType(content.MediaType).
		SetSource(content.Source).
		SetStatus(string(content.Status)).
		SetSizeBytes(content.SizeBytes).
		SetOriginalBlobHash(content.OriginalBlobHash).
		SetCreatedAt(content.CreatedAt).
		SetUpdatedAt(content.UpdatedAt)

	if content.ProcessedBlobHash != nil {
		create = create.SetProcessedBlobHash(*content.ProcessedBlobHash)
	}
	if content.ContentSummary != nil {
		create = create.SetContentSummary(*content.ContentSummary)
	}
	if content.Keywords != nil {
		create = create.SetKeywords(content.Keywords)
	}
	if content.DeletedAt != nil {
		create = create.SetDeletedAt(*content.DeletedAt)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return err
	}

	// NOTE: Graph repo is disabled (Neo4j removed)
	// Create corresponding graph node right after content source is created
	// if r.graphRepo != nil {
	// 	err = r.graphRepo.CreateContentNode(ctx, content.ID, content.SpaceID, content.Title, content.ContentSummary)
	// 	if err != nil {
	// 		// Log error but don't fail the operation
	// 		slog.Error("Failed to create content node in graph", "ContentID", content.ID, "SpaceID", content.SpaceID, "error", err)
	// 	}
	// }

	return nil
}

func (r *contentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ContentSource, error) {
	content, err := r.client.ContentSource.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.ContentSource{
		ID:                content.ID,
		SpaceID:           content.SpaceID,
		OwnerID:           content.OwnerID,
		Title:             content.Title,
		MediaType:         content.MediaType,
		Source:            content.Source,
		Status:            domain.ContentStatus(content.Status),
		SizeBytes:         content.SizeBytes,
		OriginalBlobHash:  content.OriginalBlobHash,
		ProcessedBlobHash: content.ProcessedBlobHash,
		ContentSummary:    content.ContentSummary,
		Keywords:          content.Keywords,
		CreatedAt:         content.CreatedAt,
		UpdatedAt:         content.UpdatedAt,
		DeletedAt:         content.DeletedAt,
	}, nil
}

func (r *contentRepository) List(ctx context.Context, filter domain.ContentSourceFilter) ([]*domain.ContentSource, error) {
	query := r.client.ContentSource.Query()

	if filter.SpaceID != uuid.Nil {
		query = query.Where(contentsource.SpaceID(filter.SpaceID))
	}

	if filter.OwnerID != uuid.Nil {
		query = query.Where(contentsource.OwnerID(filter.OwnerID))
	}

	// Handle title filter
	if filter.Title != "" {
		query = query.Where(contentsource.TitleContains(filter.Title))
	}

	// Handle general query (searches title and content_summary)
	if filter.Query != "" {
		query = query.Where(
			contentsource.Or(
				contentsource.TitleContains(filter.Query),
				contentsource.ContentSummaryContains(filter.Query),
			),
		)
	}

	// Handle media type filter
	if filter.MediaType != "" {
		query = query.Where(contentsource.MediaType(filter.MediaType))
	}

	// Handle keywords filter
	if len(filter.Keywords) > 0 {
		for _, keyword := range filter.Keywords {
			query = query.Where(
				contentsource.Or(
					contentsource.TitleContains(keyword),
					contentsource.ContentSummaryContains(keyword),
				),
			)
		}
	}

	// Handle status filter
	if filter.Status != "" {
		query = query.Where(contentsource.Status(string(filter.Status)))
	}

	// Handle datetime filters
	if filter.CreatedAfter != nil {
		query = query.Where(contentsource.CreatedAtGTE(*filter.CreatedAfter))
	}
	if filter.CreatedBefore != nil {
		query = query.Where(contentsource.CreatedAtLTE(*filter.CreatedBefore))
	}
	if filter.UpdatedAfter != nil {
		query = query.Where(contentsource.UpdatedAtGTE(*filter.UpdatedAfter))
	}
	if filter.UpdatedBefore != nil {
		query = query.Where(contentsource.UpdatedAtLTE(*filter.UpdatedBefore))
	}

	query = query.Order(ent.Desc(contentsource.FieldCreatedAt))

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	contents, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.ContentSource, len(contents))
	for i, c := range contents {
		result[i] = &domain.ContentSource{
			ID:                c.ID,
			SpaceID:           c.SpaceID,
			OwnerID:           c.OwnerID,
			Title:             c.Title,
			MediaType:         c.MediaType,
			Source:            c.Source,
			Status:            domain.ContentStatus(c.Status),
			SizeBytes:         c.SizeBytes,
			OriginalBlobHash:  c.OriginalBlobHash,
			ProcessedBlobHash: c.ProcessedBlobHash,
			ContentSummary:    c.ContentSummary,
			Keywords:          c.Keywords,
			CreatedAt:         c.CreatedAt,
			UpdatedAt:         c.UpdatedAt,
			DeletedAt:         c.DeletedAt,
		}
	}

	return result, nil
}

func (r *contentRepository) Update(ctx context.Context, content *domain.ContentSource) error {
	content.UpdatedAt = time.Now()

	update := r.client.ContentSource.UpdateOneID(content.ID)

	if content.Title != "" {
		update = update.SetTitle(content.Title)
	}

	if content.MediaType != "" {
		update = update.SetMediaType(content.MediaType)
	}

	if content.Source != "" {
		update = update.SetSource(content.Source)
	}

	if content.Status != "" {
		update = update.SetStatus(string(content.Status))
	}

	if content.OriginalBlobHash != "" {
		update = update.SetOriginalBlobHash(content.OriginalBlobHash)
	}

	if content.ProcessedBlobHash != nil {
		update = update.SetProcessedBlobHash(*content.ProcessedBlobHash)
	}

	if content.ContentSummary != nil {
		update = update.SetContentSummary(*content.ContentSummary)
	}

	if content.Keywords != nil {
		update = update.SetKeywords(content.Keywords)
	}

	update = update.SetUpdatedAt(content.UpdatedAt)

	_, err := update.Save(ctx)
	return err
}

func (r *contentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ContentStatus, processedBlobHash string) error {
	update := r.client.ContentSource.UpdateOneID(id).
		SetStatus(string(status)).
		SetUpdatedAt(time.Now())

	if processedBlobHash != "" {
		update = update.SetProcessedBlobHash(processedBlobHash)
	}

	_, err := update.Save(ctx)
	return err
}

func (r *contentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// First get the content to retrieve SpaceID for graph cleanup
	// content, err := r.GetByID(ctx, id) // NOTE: replace with this when graph repo is enabled
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the content source from the database
	err = r.client.ContentSource.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	// NOTE: Graph repo is disabled (Neo4j removed)
	// Clean up the corresponding graph node
	// if r.graphRepo != nil {
	// 	err = r.graphRepo.DeleteContentNode(ctx, content.ID, content.SpaceID)
	// 	if err != nil {
	// 		// Log error but don't fail the operation
	// 		slog.Error("Failed to delete content node in graph", "ContentID", content.ID, "SpaceID", content.SpaceID, "error", err)
	// 	}
	// }

	return nil
}

func (r *contentRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.client.ContentSource.Query().Where(contentsource.ID(id)).Exist(ctx)
}

func (r *contentRepository) CountBySpace(ctx context.Context, spaceID uuid.UUID) (int64, error) {
	count, err := r.client.ContentSource.Query().
		Where(contentsource.SpaceID(spaceID)).
		Count(ctx)
	return int64(count), err
}
