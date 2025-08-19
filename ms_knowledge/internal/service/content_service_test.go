package service

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

func TestCreateUploadURL_Validation(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, "", log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	spaceID := uuid.New()

	if _, _, err := svc.CreateUploadURL(ctx, uuid.Nil, "file.txt", "text/plain", 10, "title"); err == nil {
		t.Fatalf("expected error for empty space id")
	}
	if _, _, err := svc.CreateUploadURL(ctx, spaceID, "", "text/plain", 10, "title"); err == nil {
		t.Fatalf("expected error for empty filename")
	}
	if _, _, err := svc.CreateUploadURL(ctx, spaceID, "file.txt", "", 10, "title"); err == nil {
		t.Fatalf("expected error for empty mime type")
	}
	// Space not found
	if _, _, err := svc.CreateUploadURL(ctx, spaceID, "file.txt", "text/plain", 10, "title"); err == nil {
		t.Fatalf("expected error for space not found")
	}
}

func TestCreateUploadURL_Success(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, "", log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: uuid.New(), CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctx, space)

	content, url, err := svc.CreateUploadURL(ctx, space.ID, "note.pdf", "application/pdf", 123, "  Report  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content.ID == uuid.Nil {
		t.Fatalf("expected content id to be set")
	}
	if content.Status != domain.ContentStatusUploading {
		t.Fatalf("expected status UPLOADING, got %s", content.Status)
	}
	if content.Title != "Report" {
		t.Fatalf("expected trimmed title, got %q", content.Title)
	}
	if !strings.Contains(url, "spaces/") || !strings.Contains(url, "note.pdf") {
		t.Fatalf("unexpected upload url: %s", url)
	}
}

func TestConfirmUpload_UpdatesStatusAndHash(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, "", log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: uuid.New(), CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctx, space)

	item := &domain.ContentSource{ID: uuid.New(), SpaceID: space.ID, Title: "Doc", MediaType: "text/plain", Source: "f", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctx, item)

	updated, err := svc.ConfirmUpload(ctx, item.ID, "original-hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != domain.ContentStatusUploaded {
		t.Fatalf("expected status UPLOADED, got %s", updated.Status)
	}
	if updated.OriginalBlobHash != "original-hash" {
		t.Fatalf("expected original blob hash to be set")
	}
}

func TestUpdateContentSourceStatus(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, "", log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: uuid.New(), CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctx, space)

	item := &domain.ContentSource{ID: uuid.New(), SpaceID: space.ID, Title: "Doc", MediaType: "text/plain", Source: "f", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctx, item)

	res, err := svc.UpdateContentSourceStatus(ctx, item.ID, domain.ContentStatusProcessed, "processed-hash", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != domain.ContentStatusProcessed {
		t.Fatalf("expected status PROCESSED, got %s", res.Status)
	}
	if res.ProcessedBlobHash == nil || *res.ProcessedBlobHash != "processed-hash" {
		t.Fatalf("expected processed hash to be set")
	}
}
