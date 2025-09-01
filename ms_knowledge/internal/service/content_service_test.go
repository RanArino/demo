package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"demo/ms_knowledge/internal/config"
	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

// erroringStorage implements StorageService and returns an error on DeleteObject
type erroringStorage struct{ mockStorage }

func (e *erroringStorage) DeleteObject(bucket, key string) error { return fmt.Errorf("boom") }

func TestCreateUploadURL_Validation(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	cfg := &config.Config{}
	cfg.R2.BucketContentSourceName = "test-source"
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, cfg, log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	spaceID := uuid.New()

	if _, _, _, _, err := svc.CreateUploadURL(ctx, uuid.Nil, "file.txt", "text/plain", 10, "title"); err == nil {
		t.Fatalf("expected error for empty space id")
	}
	if _, _, _, _, err := svc.CreateUploadURL(ctx, spaceID, "", "text/plain", 10, "title"); err == nil {
		t.Fatalf("expected error for empty filename")
	}
	if _, _, _, _, err := svc.CreateUploadURL(ctx, spaceID, "file.txt", "", 10, "title"); err == nil {
		t.Fatalf("expected error for empty mime type")
	}
	// Space not found
	if _, _, _, _, err := svc.CreateUploadURL(ctx, spaceID, "file.txt", "text/plain", 10, "title"); err == nil {
		t.Fatalf("expected error for space not found")
	}
}

func TestCreateUploadURL_Success(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	cfg := &config.Config{}
	cfg.R2.BucketContentSourceName = "test-source"
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, cfg, log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	ownerID := uuid.New()
	ctxOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: ownerID, CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctxOwner, space)

	content, url, _, _, err := svc.CreateUploadURL(ctxOwner, space.ID, "note.pdf", "application/pdf", 123, "  Report  ")
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
	cfg := &config.Config{}
	cfg.R2.BucketContentSourceName = "test-source"
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, cfg, log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	ownerID := uuid.New()
	ctxOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: ownerID, CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctxOwner, space)

	item := &domain.ContentSource{ID: uuid.New(), SpaceID: space.ID, OwnerID: ownerID, Title: "Doc", MediaType: "text/plain", Source: "f", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctxOwner, item)

	updated, err := svc.ConfirmUpload(ctxOwner, item.ID, "original-hash")
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
	cfg := &config.Config{}
	cfg.R2.BucketContentSourceName = "test-source"
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, cfg, log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	ownerID := uuid.New()
	ctxOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: ownerID, CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctxOwner, space)

	item := &domain.ContentSource{ID: uuid.New(), SpaceID: space.ID, OwnerID: ownerID, Title: "Doc", MediaType: "text/plain", Source: "f", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctxOwner, item)

	res, err := svc.UpdateContentSourceStatus(ctxOwner, item.ID, domain.ContentStatusProcessed, "processed-hash", "")
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

func TestDeleteContentSource_DeletesR2AndDB(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	cfg := &config.Config{}
	cfg.R2.BucketContentSourceName = "knowledge-source"
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, cfg, log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	ownerID := uuid.New()
	ctxOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: ownerID, CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctxOwner, space)

	item := &domain.ContentSource{ID: uuid.New(), SpaceID: space.ID, OwnerID: ownerID, Title: "Doc", MediaType: "text/plain", Source: "file.txt", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctxOwner, item)

	if err := svc.DeleteContentSource(ctxOwner, item.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exists, _ := contentRepo.Exists(ctxOwner, item.ID); exists {
		t.Fatalf("expected DB row to be deleted")
	}

	if len(storage.deleted) != 2 {
		t.Fatalf("expected two delete calls to storage (original and processed), got %d", len(storage.deleted))
	}
	// Check that both original and processed files are deleted
	expectedOrigKey := ownerID.String() + "/spaces/" + space.ID.String() + "/content/" + item.ID.String() + "/file.txt"
	expectedProcKey := ownerID.String() + "/spaces/" + space.ID.String() + "/content/" + item.ID.String() + "/file.md"
	
	keys := []string{storage.deleted[0].key, storage.deleted[1].key}
	if !contains(keys, expectedOrigKey) || !contains(keys, expectedProcKey) {
		t.Fatalf("expected keys %s and %s, got %v", expectedOrigKey, expectedProcKey, keys)
	}
}

func TestDeleteContentSource_ObjectDeleteFailsButDBStillRemoved(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	errStorage := &erroringStorage{}
	cfg := &config.Config{}
	cfg.R2.BucketContentSourceName = "knowledge-source"
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, errStorage, nil, cfg, log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	ownerID := uuid.New()
	ctxOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: ownerID, CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctxOwner, space)

	item := &domain.ContentSource{ID: uuid.New(), SpaceID: space.ID, OwnerID: ownerID, Title: "Doc", MediaType: "text/plain", Source: "file.txt", Status: domain.ContentStatusUploaded}
	_ = contentRepo.Create(ctxOwner, item)

	if err := svc.DeleteContentSource(ctxOwner, item.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists, _ := contentRepo.Exists(ctxOwner, item.ID); exists {
		t.Fatalf("expected DB row to be deleted even if object delete fails")
	}
}

func TestCreateUploadURL_DefaultsTitleToFilename(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	cfg := &config.Config{}
	cfg.R2.BucketContentSourceName = "test-source"
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, cfg, log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	ownerID := uuid.New()
	ctxOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: ownerID, CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctxOwner, space)

	content, _, _, _, err := svc.CreateUploadURL(ctxOwner, space.ID, "file-name.txt", "text/plain", 1, "   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content.Title != "file-name.txt" {
		t.Fatalf("expected title to default to filename, got %q", content.Title)
	}
}

func TestUpdateContentSource_PartialUpdates(t *testing.T) {
	contentRepo := newMockContentRepo()
	spaceRepo := newMockSpaceRepo()
	graphRepo := newMockGraphRepo()
	storage := &mockStorage{}
	cfg := &config.Config{}
	cfg.R2.BucketContentSourceName = "test-source"
	svc := NewContentService(contentRepo, spaceRepo, graphRepo, storage, nil, cfg, log.New(os.Stdout, "", log.LstdFlags))

	ctx := context.Background()
	ownerID := uuid.New()
	ctxOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())
	space := &domain.Space{ID: uuid.New(), Title: "s", OwnerID: ownerID, CreatedAt: time.Now(), LastUpdatedAt: time.Now()}
	_ = spaceRepo.Create(ctxOwner, space)

	item := &domain.ContentSource{ID: uuid.New(), SpaceID: space.ID, OwnerID: ownerID, Title: "Old", MediaType: "text/plain", Source: "file.txt", Status: domain.ContentStatusUploaded, Keywords: []string{"a", "b"}}
	_ = contentRepo.Create(ctxOwner, item)

	// Update title only
	newTitle := "New Title"
	updated, err := svc.UpdateContentSource(ctxOwner, item.ID, &newTitle, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != newTitle {
		t.Fatalf("expected title updated, got %q", updated.Title)
	}
	if len(updated.Keywords) != 2 {
		t.Fatalf("expected keywords unchanged, got %v", updated.Keywords)
	}

	// Update keywords only (clear then set)
	empty := []string{}
	updated, err = svc.UpdateContentSource(ctxOwner, item.ID, nil, &empty)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updated.Keywords) != 0 {
		t.Fatalf("expected keywords cleared, got %v", updated.Keywords)
	}
	ks := []string{"x", "y"}
	updated, err = svc.UpdateContentSource(ctxOwner, item.ID, nil, &ks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := strings.Join(updated.Keywords, ","), "x,y"; got != want {
		t.Fatalf("expected keywords %q, got %q", want, got)
	}
}
// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}