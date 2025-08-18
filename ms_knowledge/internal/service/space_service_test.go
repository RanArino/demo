package service

import (
	"context"
	"testing"
	"time"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

func TestCreateSpace_ValidationErrors(t *testing.T) {
	spaceRepo := newMockSpaceRepo()
	contentRepo := newMockContentRepo()
	graphRepo := newMockGraphRepo()
	svc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()

	if _, err := svc.CreateSpace(ctx, "", "desc", uuid.New().String()); err == nil {
		t.Fatalf("expected error for empty title")
	}
	if _, err := svc.CreateSpace(ctx, "title", "desc", ""); err == nil {
		t.Fatalf("expected error for empty owner id")
	}
	if _, err := svc.CreateSpace(ctx, "title", "desc", "not-a-uuid"); err == nil {
		t.Fatalf("expected error for invalid owner id format")
	}
}

func TestCreateSpace_Success(t *testing.T) {
	spaceRepo := newMockSpaceRepo()
	contentRepo := newMockContentRepo()
	graphRepo := newMockGraphRepo()
	svc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()
	owner := uuid.New().String()

	sp, err := svc.CreateSpace(ctx, "  My Space  ", "  Desc  ", owner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sp.ID == uuid.Nil {
		t.Fatalf("expected id to be set")
	}
	if sp.Title != "My Space" {
		t.Fatalf("expected trimmed title, got %q", sp.Title)
	}
}

func TestDeleteSpace_BlocksWhenContentExists(t *testing.T) {
	spaceRepo := newMockSpaceRepo()
	contentRepo := newMockContentRepo()
	graphRepo := newMockGraphRepo()
	svc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()
	// Prepare space and content
	sp := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: uuid.New(), CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctx, sp)
	_ = contentRepo.Create(ctx, &domain.ContentSource{ID: uuid.New(), SpaceID: sp.ID, Title: "c", MediaType: "text/plain", Source: "f", Status: domain.ContentStatusUploading})

	if err := svc.DeleteSpace(ctx, sp.ID, false, false); err == nil {
		t.Fatalf("expected error when space has content")
	}
}

func TestDeleteSpace_ForceBypassContentCheck(t *testing.T) {
	spaceRepo := newMockSpaceRepo()
	contentRepo := newMockContentRepo()
	graphRepo := newMockGraphRepo()
	svc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()
	sp := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: uuid.New(), CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctx, sp)
	_ = contentRepo.Create(ctx, &domain.ContentSource{ID: uuid.New(), SpaceID: sp.ID, Title: "c", MediaType: "text/plain", Source: "f", Status: domain.ContentStatusUploading})

	if err := svc.DeleteSpace(ctx, sp.ID, true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spaceRepo.deletedCalls) != 1 || spaceRepo.deletedCalls[0].id != sp.ID || !spaceRepo.deletedCalls[0].hard {
		t.Fatalf("expected delete to be called with hard=true for space %s", sp.ID)
	}
}

func TestSearchSpaces_AddsStats(t *testing.T) {
	spaceRepo := newMockSpaceRepo()
	contentRepo := newMockContentRepo()
	graphRepo := newMockGraphRepo()
	svc := NewSpaceService(spaceRepo, contentRepo, graphRepo)

	ctx := context.Background()
	owner := uuid.New()
	sp := &domain.Space{ID: uuid.New(), Title: "Alpha", Description: "d", OwnerID: owner, CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctx, sp)
	// One content in this space
	_ = contentRepo.Create(ctx, &domain.ContentSource{ID: uuid.New(), SpaceID: sp.ID, Title: "Doc", MediaType: "text/plain", Source: "f", Status: domain.ContentStatusUploaded})

	spaceRepo.searchResult = []*domain.Space{sp}
	res, err := svc.SearchSpaces(ctx, "Alpha", domain.SpaceFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Stats.ContentCount != 1 {
		t.Fatalf("expected content count 1, got %d", res[0].Stats.ContentCount)
	}
}
