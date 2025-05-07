package s3

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// S3Client wraps AWS S3 SDK client
type S3Client struct {
	client        *s3.Client
	config        *document.S3Config
	presignClient *s3.PresignClient
	logger        *slog.Logger
}

// NewS3Client creates a new S3 client
func NewS3Client(cfg *document.S3Config, logger *slog.Logger) (*S3Client, error) {
	// Create AWS SDK configuration
	awsConfig, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Create S3 client
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
	})

	// Create presign client for generating presigned URLs
	presignClient := s3.NewPresignClient(client)

	return &S3Client{
		client:        client,
		config:        cfg,
		presignClient: presignClient,
		logger:        logger,
	}, nil
}

// loadAWSConfig creates AWS SDK configuration from S3Config
func loadAWSConfig(cfg *document.S3Config) (aws.Config, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var optFns []func(*config.LoadOptions) error

	// Configure region
	optFns = append(optFns, config.WithRegion(cfg.Region))

	// Configure credentials if provided (for local development)
	if !cfg.UseIAMRoleCredentials && cfg.AccessKey != "" && cfg.SecretKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	// Configure endpoint for S3-compatible services
	if cfg.Endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				HostnameImmutable: true,
				SigningRegion:     cfg.Region,
			}, nil
		})
		optFns = append(optFns, config.WithEndpointResolverWithOptions(customResolver))
	}

	// Configure HTTP client
	if cfg.DisableSSL {
		optFns = append(optFns, config.WithHTTPClient(makeHTTPClient(true)))
	}

	// Load AWS configuration
	return config.LoadDefaultConfig(ctx, optFns...)
}

// makeHTTPClient creates an HTTP client with custom settings
func makeHTTPClient(disableSSL bool) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: disableSSL,
		},
	}
	return &http.Client{Transport: tr, Timeout: 60 * time.Second}
}

// GetClient returns the underlying S3 client
func (c *S3Client) GetClient() *s3.Client {
	return c.client
}

// GetPresignClient returns the presign client for generating presigned URLs
func (c *S3Client) GetPresignClient() *s3.PresignClient {
	return c.presignClient
}

// GetConfig returns the S3 configuration
func (c *S3Client) GetConfig() *document.S3Config {
	return c.config
}

// HeadBucket checks if a bucket exists and is accessible
func (c *S3Client) HeadBucket(ctx context.Context, bucketName string) error {
	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

// CreateBucket creates a new S3 bucket if it doesn't exist
func (c *S3Client) CreateBucket(ctx context.Context, bucketName string) error {
	_, err := c.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(c.config.Region),
		},
	})
	return err
}
