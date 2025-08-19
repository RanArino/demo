package r2

import (
	"context"
	"time"
)

// Adapter implements ms_knowledge/internal/service.StorageService on top of the R2 client.
type Adapter struct {
	client *Client
}

func NewAdapter(client *Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) GeneratePresignedUploadURL(bucket, key string, expires time.Duration) (string, error) {
	return a.client.GeneratePresignedUploadURL(context.Background(), bucket, key, expires, "")
}

func (a *Adapter) CalculateSHA256(data []byte) string {
	return CalculateSHA256(data)
}


