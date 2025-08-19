//go:build integration
// +build integration

package r2

import (
	"context"
	"demo/ms_knowledge/internal/config"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestR2Upload uploads the repository's test PDF to the configured R2 source bucket
// and verifies the object can be downloaded back.
func TestR2Upload(t *testing.T) {
	// Load env from ms_knowledge/.env.local if present when running locally
	_, _ = config.LoadForDevelopment()

	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET_SOURCE_NAME")
	region := os.Getenv("R2_REGION")
	if region == "" {
		region = "auto"
	}
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("R2 env vars not set; skipping integration test (need R2_ENDPOINT, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_SOURCE_NAME)")
	}

	client, err := NewClient(context.Background(), Config{
		Endpoint:        endpoint,
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		UsePathStyle:    true,
	})
	if err != nil {
		t.Fatalf("failed creating R2 client: %v", err)
	}

	// Read the test PDF from repo: ms_knowledge/test.pdf
	pwd, _ := os.Getwd()
	// This test file lives in ms_knowledge/internal/storage/r2, so go up three levels to ms_knowledge
	pdfPath := filepath.Clean(filepath.Join(pwd, "../../../test.pdf"))
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("failed reading test PDF at %s: %v", pdfPath, err)
	}

	key := filepath.ToSlash(filepath.Join("integration-tests", uuid.New().String()+".pdf"))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.Upload(ctx, bucket, key, data, "application/pdf"); err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	got, err := client.Download(ctx, bucket, key)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}
	if len(got) != len(data) {
		t.Fatalf("size mismatch: uploaded %d bytes, downloaded %d bytes", len(data), len(got))
	}

	t.Logf("uploaded to bucket %q key %q (%d bytes)", bucket, key, len(got))
}
