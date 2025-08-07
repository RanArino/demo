package secrets

import (
	"context"
)

// SecretManager defines the interface for retrieving secrets from various sources.
type SecretManager interface {
	// GetSecret retrieves a secret by key
	GetSecret(ctx context.Context, key string) (string, error)
	
	// GetSecrets retrieves multiple secrets by keys
	GetSecrets(ctx context.Context, keys []string) (map[string]string, error)
	
	// RefreshSecrets refreshes cached secrets (if applicable)
	RefreshSecrets(ctx context.Context) error
	
	// Close closes any underlying connections
	Close() error
}

// SecretProvider represents the type of secret provider
type SecretProvider string

const (
	ProviderEnv    SecretProvider = "env"
	ProviderAWS    SecretProvider = "aws"
	ProviderGoogle SecretProvider = "google"
)

// Config holds configuration for secret managers
type Config struct {
	Provider SecretProvider `json:"provider"`
	
	// AWS specific configuration
	AWSRegion    string `json:"aws_region,omitempty"`
	AWSSecretArn string `json:"aws_secret_arn,omitempty"`
	
	// Google specific configuration
	GoogleProjectID string `json:"google_project_id,omitempty"`
	GoogleSecretID  string `json:"google_secret_id,omitempty"`
	
	// Environment specific configuration
	EnvPrefix string `json:"env_prefix,omitempty"`
}