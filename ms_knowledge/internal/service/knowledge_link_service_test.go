package service

import (
	"context"
	"testing"
	"time"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

func TestCreateKnowledgeLink_Validation(t *testing.T) {
	graphRepo := newMockGraphRepo()
	contentRepo := newMockContentRepo()
	svc := NewKnowledgeLinkService(graphRepo, contentRepo)
	ctx := context.Background()

	if _, err := svc.CreateKnowledgeLink(ctx, uuid.Nil, uuid.New(), domain.RelationTypeReferences, 1.0); err == nil {
		t.Fatalf("expected error for empty from id")
	}
	if _, err := svc.CreateKnowledgeLink(ctx, uuid.New(), uuid.Nil, domain.RelationTypeReferences, 1.0); err == nil {
		t.Fatalf("expected error for empty to id")
	}
	common := uuid.New()
	if _, err := svc.CreateKnowledgeLink(ctx, common, common, domain.RelationTypeReferences, 1.0); err == nil {
		t.Fatalf("expected error for self-link")
	}
}

func TestCreateKnowledgeLink_DefaultsAndSameSpace(t *testing.T) {
	graphRepo := newMockGraphRepo()
	contentRepo := newMockContentRepo()
	svc := NewKnowledgeLinkService(graphRepo, contentRepo)
	ctx := context.Background()

	spaceID := uuid.New()
	from := &domain.ContentSource{ID: uuid.New(), SpaceID: spaceID, Title: "A", MediaType: "text/plain", Source: "a", Status: domain.ContentStatusUploaded}
	to := &domain.ContentSource{ID: uuid.New(), SpaceID: spaceID, Title: "B", MediaType: "text/plain", Source: "b", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctx, from)
	_ = contentRepo.Create(ctx, to)

	link, err := svc.CreateKnowledgeLink(ctx, from.ID, to.ID, "", 2.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.RelationType != domain.RelationTypeReferences {
		t.Fatalf("expected default relation type REFERENCES, got %s", link.RelationType)
	}
	if link.Weight == nil || *link.Weight != 1.0 {
		t.Fatalf("expected weight to be clamped to 1.0")
	}
}

func TestCreateKnowledgeLink_CrossSpaceForbidden(t *testing.T) {
	graphRepo := newMockGraphRepo()
	contentRepo := newMockContentRepo()
	svc := NewKnowledgeLinkService(graphRepo, contentRepo)
	ctx := context.Background()

	from := &domain.ContentSource{ID: uuid.New(), SpaceID: uuid.New(), Title: "A", MediaType: "text/plain", Source: "a", Status: domain.ContentStatusUploaded}
	to := &domain.ContentSource{ID: uuid.New(), SpaceID: uuid.New(), Title: "B", MediaType: "text/plain", Source: "b", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctx, from)
	_ = contentRepo.Create(ctx, to)

	if _, err := svc.CreateKnowledgeLink(ctx, from.ID, to.ID, domain.RelationTypeRelated, 0.5); err == nil {
		t.Fatalf("expected error for cross-space link")
	}
}

func TestBacklinksAndCount(t *testing.T) {
	graphRepo := newMockGraphRepo()
	contentRepo := newMockContentRepo()
	svc := NewKnowledgeLinkService(graphRepo, contentRepo)
	ctx := context.Background()

	spaceID := uuid.New()
	a := &domain.ContentSource{ID: uuid.New(), SpaceID: spaceID, Title: "A", MediaType: "text/plain", Source: "a", Status: domain.ContentStatusUploaded}
	b := &domain.ContentSource{ID: uuid.New(), SpaceID: spaceID, Title: "B", MediaType: "text/plain", Source: "b", Status: domain.ContentStatusUploaded}
	c := &domain.ContentSource{ID: uuid.New(), SpaceID: spaceID, Title: "C", MediaType: "text/plain", Source: "c", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctx, a)
	_ = contentRepo.Create(ctx, b)
	_ = contentRepo.Create(ctx, c)

	// Create two inbound links to B
	_ = graphRepo.CreateKnowledgeLink(ctx, &domain.KnowledgeLink{ID: uuid.New(), FromContentID: a.ID, ToContentID: b.ID, RelationType: domain.RelationTypeReferences, Weight: floatPtr(1.0), CreatedAt: time.Now(), UpdatedAt: time.Now()})
	_ = graphRepo.CreateKnowledgeLink(ctx, &domain.KnowledgeLink{ID: uuid.New(), FromContentID: c.ID, ToContentID: b.ID, RelationType: domain.RelationTypeRelated, Weight: floatPtr(0.8), CreatedAt: time.Now(), UpdatedAt: time.Now()})

	backlinks, err := svc.GetBacklinks(ctx, b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(backlinks) != 2 {
		t.Fatalf("expected 2 backlinks, got %d", len(backlinks))
	}

	count, err := svc.CountLinksByContent(ctx, b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected link count 2, got %d", count)
	}
}

func TestCRUDPassThrough(t *testing.T) {
	graphRepo := newMockGraphRepo()
	contentRepo := newMockContentRepo()
	svc := NewKnowledgeLinkService(graphRepo, contentRepo)
	ctx := context.Background()

	from := uuid.New()
	to := uuid.New()
	weight := 0.7
	link := &domain.KnowledgeLink{ID: uuid.New(), FromContentID: from, ToContentID: to, RelationType: domain.RelationTypeRelated, Weight: &weight, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := graphRepo.CreateKnowledgeLink(ctx, link); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	// Get
	got, err := svc.GetKnowledgeLink(ctx, link.ID)
	if err != nil || got.ID != link.ID {
		t.Fatalf("get failed: %v", err)
	}

	// Update
	newWeight := 0.5
	_, _ = svc.UpdateKnowledgeLink(ctx, link.ID, domain.RelationTypeContains, newWeight)
	// Verify
	stored, _ := graphRepo.GetKnowledgeLink(ctx, link.ID)
	if stored.RelationType != domain.RelationTypeContains || stored.Weight == nil || *stored.Weight != newWeight {
		t.Fatalf("update not applied")
	}

	// Delete
	if err := svc.DeleteKnowledgeLink(ctx, link.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := graphRepo.GetKnowledgeLink(ctx, link.ID); err == nil {
		t.Fatalf("expected link to be deleted")
	}
}

func floatPtr(v float64) *float64 { return &v }
