package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

// mockSpaceRepo is an in-memory implementation of domain.SpaceRepository for tests
type mockSpaceRepo struct {
	spaces        map[uuid.UUID]*domain.Space
	withStats     map[uuid.UUID]*domain.SpaceWithStats
	list          []*domain.Space
	listWithStats []*domain.SpaceWithStats
	searchResult  []*domain.Space
	deletedCalls  []struct {
		id   uuid.UUID
		hard bool
	}
}

func newMockSpaceRepo() *mockSpaceRepo {
	return &mockSpaceRepo{
		spaces:    make(map[uuid.UUID]*domain.Space),
		withStats: make(map[uuid.UUID]*domain.SpaceWithStats),
	}
}

func (m *mockSpaceRepo) Create(ctx context.Context, space *domain.Space) error {
	if space.ID == uuid.Nil {
		space.ID = uuid.New()
	}
	if space.CreatedAt.IsZero() {
		space.CreatedAt = time.Now()
	}
	if space.LastUpdatedAt.IsZero() {
		space.LastUpdatedAt = time.Now()
	}
	m.spaces[space.ID] = space
	return nil
}

func (m *mockSpaceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Space, error) {
	sp, ok := m.spaces[id]
	if !ok {
		return nil, fmt.Errorf("space not found")
	}
	return sp, nil
}

func (m *mockSpaceRepo) GetWithStats(ctx context.Context, id uuid.UUID) (*domain.SpaceWithStats, error) {
	if ws, ok := m.withStats[id]; ok {
		return ws, nil
	}
	sp, err := m.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.SpaceWithStats{Space: *sp, Stats: domain.SpaceStats{}}, nil
}

func (m *mockSpaceRepo) List(ctx context.Context, filter domain.SpaceFilter) ([]*domain.Space, error) {
	if m.list != nil {
		return m.list, nil
	}
	spaces := make([]*domain.Space, 0, len(m.spaces))
	for _, s := range m.spaces {
		spaces = append(spaces, s)
	}
	sort.Slice(spaces, func(i, j int) bool { return spaces[i].CreatedAt.After(spaces[j].CreatedAt) })
	return spaces, nil
}

func (m *mockSpaceRepo) ListWithStats(ctx context.Context, filter domain.SpaceFilter) ([]*domain.SpaceWithStats, error) {
	if m.listWithStats != nil {
		return m.listWithStats, nil
	}
	list, _ := m.List(ctx, filter)
	res := make([]*domain.SpaceWithStats, 0, len(list))
	for _, s := range list {
		res = append(res, &domain.SpaceWithStats{Space: *s, Stats: domain.SpaceStats{}})
	}
	return res, nil
}

func (m *mockSpaceRepo) Update(ctx context.Context, space *domain.Space) error {
	if space == nil || space.ID == uuid.Nil {
		return fmt.Errorf("invalid space")
	}
	m.spaces[space.ID] = space
	return nil
}

func (m *mockSpaceRepo) Delete(ctx context.Context, id uuid.UUID, hardDelete bool) error {
	m.deletedCalls = append(m.deletedCalls, struct {
		id   uuid.UUID
		hard bool
	}{id: id, hard: hardDelete})
	delete(m.spaces, id)
	return nil
}

func (m *mockSpaceRepo) Search(ctx context.Context, query string, filter domain.SpaceFilter) ([]*domain.Space, error) {
	if m.searchResult != nil {
		return m.searchResult, nil
	}
	return m.List(ctx, filter)
}

func (m *mockSpaceRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.spaces[id]
	return ok, nil
}

// mockContentRepo is an in-memory implementation of domain.ContentRepository for tests
type mockContentRepo struct {
	contents map[uuid.UUID]*domain.ContentSource
}

func newMockContentRepo() *mockContentRepo {
	return &mockContentRepo{contents: make(map[uuid.UUID]*domain.ContentSource)}
}

func (m *mockContentRepo) Create(ctx context.Context, content *domain.ContentSource) error {
	if content.ID == uuid.Nil {
		content.ID = uuid.New()
	}
	if content.CreatedAt.IsZero() {
		content.CreatedAt = time.Now()
	}
	content.UpdatedAt = content.CreatedAt
	m.contents[content.ID] = content
	return nil
}

func (m *mockContentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ContentSource, error) {
	c, ok := m.contents[id]
	if !ok {
		return nil, fmt.Errorf("content not found")
	}
	return c, nil
}

func (m *mockContentRepo) List(ctx context.Context, filter domain.ContentSourceFilter) ([]*domain.ContentSource, error) {
	var res []*domain.ContentSource
	for _, c := range m.contents {
		if filter.SpaceID != uuid.Nil && c.SpaceID != filter.SpaceID {
			continue
		}
		res = append(res, c)
	}
	return res, nil
}

func (m *mockContentRepo) Update(ctx context.Context, content *domain.ContentSource) error {
	if content == nil || content.ID == uuid.Nil {
		return fmt.Errorf("invalid content")
	}
	content.UpdatedAt = time.Now()
	m.contents[content.ID] = content
	return nil
}

func (m *mockContentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ContentStatus, processedBlobHash string) error {
	c, ok := m.contents[id]
	if !ok {
		return fmt.Errorf("content not found")
	}
	c.Status = status
	if processedBlobHash != "" {
		c.ProcessedBlobHash = &processedBlobHash
	}
	c.UpdatedAt = time.Now()
	return nil
}

func (m *mockContentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.contents, id)
	return nil
}

func (m *mockContentRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.contents[id]
	return ok, nil
}

func (m *mockContentRepo) CountBySpace(ctx context.Context, spaceID uuid.UUID) (int64, error) {
	var count int64
	for _, c := range m.contents {
		if c.SpaceID == spaceID {
			count++
		}
	}
	return count, nil
}

// mockGraphRepo is an in-memory implementation of domain.GraphRepository for tests
type mockGraphRepo struct {
	links map[uuid.UUID]*domain.KnowledgeLink
}

func newMockGraphRepo() *mockGraphRepo {
	return &mockGraphRepo{links: make(map[uuid.UUID]*domain.KnowledgeLink)}
}

func (m *mockGraphRepo) CreateKnowledgeLink(ctx context.Context, link *domain.KnowledgeLink) error {
	if link.FromContentID == link.ToContentID {
		return fmt.Errorf("cannot create self-link")
	}
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}
	if link.CreatedAt.IsZero() {
		link.CreatedAt = time.Now()
	}
	link.UpdatedAt = link.CreatedAt
	m.links[link.ID] = link
	return nil
}

func (m *mockGraphRepo) GetKnowledgeLink(ctx context.Context, id uuid.UUID) (*domain.KnowledgeLink, error) {
	l, ok := m.links[id]
	if !ok {
		return nil, fmt.Errorf("link not found")
	}
	return l, nil
}

func (m *mockGraphRepo) ListKnowledgeLinks(ctx context.Context, filter domain.LinkFilter) ([]*domain.KnowledgeLink, error) {
	var res []*domain.KnowledgeLink
	for _, l := range m.links {
		switch filter.Direction {
		case domain.LinkDirectionInbound:
			if l.ToContentID == filter.ContentID {
				res = append(res, l)
			}
		case domain.LinkDirectionOutbound:
			if l.FromContentID == filter.ContentID {
				res = append(res, l)
			}
		case domain.LinkDirectionBoth:
			if l.FromContentID == filter.ContentID || l.ToContentID == filter.ContentID {
				res = append(res, l)
			}
		}
	}
	return res, nil
}

func (m *mockGraphRepo) UpdateKnowledgeLink(ctx context.Context, link *domain.KnowledgeLink) error {
	_, ok := m.links[link.ID]
	if !ok {
		return fmt.Errorf("link not found")
	}
	link.UpdatedAt = time.Now()
	m.links[link.ID] = link
	return nil
}

func (m *mockGraphRepo) DeleteKnowledgeLink(ctx context.Context, id uuid.UUID) error {
	delete(m.links, id)
	return nil
}

func (m *mockGraphRepo) CreateLink(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID) error {
	weight := 1.0
	return m.CreateKnowledgeLink(ctx, &domain.KnowledgeLink{ID: uuid.New(), FromContentID: fromContentID, ToContentID: toContentID, RelationType: domain.RelationTypeReferences, Weight: &weight, CreatedAt: time.Now(), UpdatedAt: time.Now()})
}

func (m *mockGraphRepo) CreateLinkWithType(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID, relationshipType string) error {
	weight := 1.0
	return m.CreateKnowledgeLink(ctx, &domain.KnowledgeLink{ID: uuid.New(), FromContentID: fromContentID, ToContentID: toContentID, RelationType: domain.RelationType(relationshipType), Weight: &weight, CreatedAt: time.Now(), UpdatedAt: time.Now()})
}

func (m *mockGraphRepo) GetBacklinks(ctx context.Context, contentID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for _, l := range m.links {
		if l.ToContentID == contentID {
			ids = append(ids, l.FromContentID)
		}
	}
	return ids, nil
}

func (m *mockGraphRepo) CreateContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID) error {
	return nil
}
func (m *mockGraphRepo) DeleteContentNode(ctx context.Context, contentID uuid.UUID, spaceID uuid.UUID) error {
	return nil
}

func (m *mockGraphRepo) CountLinksByContent(ctx context.Context, contentID uuid.UUID) (int64, error) {
	var n int64
	for _, l := range m.links {
		if l.FromContentID == contentID || l.ToContentID == contentID {
			n++
		}
	}
	return n, nil
}

func (m *mockGraphRepo) CountLinksBySpace(ctx context.Context, spaceID uuid.UUID) (int64, error) {
	return 0, nil
}

// mockStorage implements StorageService for tests
type mockStorage struct{}

func (m *mockStorage) GeneratePresignedUploadURL(bucket, key string, expires time.Duration) (string, error) {
	return fmt.Sprintf("https://example.com/%s", key), nil
}

func (m *mockStorage) CalculateSHA256(data []byte) string {
	sha := sha256.Sum256(data)
	return hex.EncodeToString(sha[:])
}
