package infra

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"demo/ms_canvas/go_app/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Client provides methods for interacting with Cloudflare R2 storage
type R2Client struct {
	client *s3.Client
	bucket string
}

// NewR2Client creates a new R2 client with the provided configuration
func NewR2Client(cfg config.R2Config) (*R2Client, error) {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.AccountID == "" || cfg.BucketProcessed == "" {
		return nil, fmt.Errorf("R2 configuration is incomplete: endpoint, access_key_id, secret_access_key, account_id, and bucket_processed are required")
	}

	// Create static credentials provider
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")

	// Load AWS config with custom endpoint for R2
	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithCredentialsProvider(creds),
		awsConfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &R2Client{
		client: client,
		bucket: cfg.BucketProcessed,
	}, nil
}

// DownloadFile downloads a file from R2 storage and returns its contents
func (r *R2Client) DownloadFile(key string) ([]byte, error) {
	result, err := r.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from R2: %w", err)
	}
	defer result.Body.Close()

	// Read the entire file
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content: %w", err)
	}

	return content, nil
}

// UploadFile uploads data to R2 storage
func (r *R2Client) UploadFile(key string, data []byte, contentType string) error {
	_, err := r.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to put object to R2: %w", err)
	}

	return nil
}

// DeleteFile deletes a file from R2 storage
func (r *R2Client) DeleteFile(key string) error {
	_, err := r.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from R2: %w", err)
	}

	return nil
}
