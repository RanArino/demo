package document

import (
	"time"
)

// StorageProvider specifies the type of storage backend
type StorageProvider string

const (
	StorageProviderS3     StorageProvider = "s3"
	StorageProviderLocal  StorageProvider = "local"
	StorageProviderMemory StorageProvider = "memory" // For testing
)

// StorageObjectID is a unique identifier for a stored object
type StorageObjectID string

// StorageOptions defines parameters for storing an object
type StorageOptions struct {
	Bucket             string            // S3 bucket or equivalent
	ContentType        string            // MIME type
	MetadataAttributes map[string]string // Additional metadata to store
	ExpiresAfter       *time.Duration    // Optional expiry for temporary objects
	PubliclyAccessible bool              // Whether the object should be publicly accessible
	StorageClass       string            // S3 storage class or equivalent
	// ACL and other provider-specific options can be added here
}

// StorageObject represents metadata about a stored object
type StorageObject struct {
	ID           StorageObjectID   `json:"id"`
	Key          string            `json:"key"`    // Storage path/key
	Bucket       string            `json:"bucket"` // S3 bucket or equivalent
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag,omitempty"` // Used for versioning
	LastModified time.Time         `json:"last_modified"`
	Provider     StorageProvider   `json:"provider"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// StorageUploadInfo contains information for an upload
type StorageUploadInfo struct {
	PresignedURL string            `json:"presigned_url,omitempty"` // For direct client->S3 uploads
	ExpiresAt    time.Time         `json:"expires_at,omitempty"`
	FormFields   map[string]string `json:"form_fields,omitempty"` // For S3 POST policy uploads
	Key          string            `json:"key"`
	Bucket       string            `json:"bucket"`
}

// StorageDownloadInfo contains information for a download
type StorageDownloadInfo struct {
	PresignedURL string    `json:"presigned_url,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	ContentType  string    `json:"content_type"`
	Size         int64     `json:"size"`
	Filename     string    `json:"filename"`
}

// S3Config defines configuration for the S3 storage provider
type S3Config struct {
	Region                string `json:"region"`
	Bucket                string `json:"bucket"`
	Endpoint              string `json:"endpoint,omitempty"`         // For S3-compatible services
	ForcePathStyle        bool   `json:"force_path_style,omitempty"` // For S3-compatible services
	DisableSSL            bool   `json:"disable_ssl,omitempty"`      // For development
	PresignedURLDuration  int    `json:"presigned_url_duration"`     // In seconds
	AccessKey             string `json:"access_key,omitempty"`       // For local development
	SecretKey             string `json:"secret_key,omitempty"`       // For local development
	UseIAMRoleCredentials bool   `json:"use_iam_role_credentials"`   // For production
}
