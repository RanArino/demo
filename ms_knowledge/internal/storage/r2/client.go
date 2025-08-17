package r2

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config holds Cloudflare R2 configuration.
type Config struct {
	// Example: https://<account_id>.r2.cloudflarestorage.com
	Endpoint        string
	Region          string // often "auto" for R2
	AccessKeyID     string
	SecretAccessKey string
	// UsePathStyle is recommended for custom endpoints like R2.
	UsePathStyle bool
}

// Client wraps S3-compatible client for R2 and a presigner.
type Client struct {
	s3        *s3.Client
	presigner *s3.PresignClient
}

// NewClient constructs an R2 client with provided config. If Endpoint is set, it overrides the endpoint resolver.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("r2 endpoint is required")
	}
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, errors.New("r2 credentials are required")
	}
	if cfg.Region == "" {
		cfg.Region = "auto"
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		o.BaseEndpoint = aws.String(cfg.Endpoint)
	})

	return &Client{
		s3:        s3Client,
		presigner: s3.NewPresignClient(s3Client),
	}, nil
}

// Upload uploads the given bytes to R2 at bucket/key.
func (c *Client) Upload(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	return err
}

// Download fetches the object bytes from R2 at bucket/key.
func (c *Client) Download(ctx context.Context, bucket, key string) ([]byte, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}

// GeneratePresignedUploadURL creates a pre-signed PUT URL for direct client upload.
func (c *Client) GeneratePresignedUploadURL(ctx context.Context, bucket, key string, expires time.Duration, contentType string) (string, error) {
	if expires <= 0 {
		expires = 15 * time.Minute
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	req, err := c.presigner.PresignPutObject(ctx, input, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// CalculateSHA256 returns the hex-encoded sha256 of data.
func CalculateSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
