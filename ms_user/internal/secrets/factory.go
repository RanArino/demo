package secrets

import (
	"context"
	"fmt"
	"os"
)

// NewSecretManager creates a new secret manager based on the provider type
func NewSecretManager(ctx context.Context, config Config) (SecretManager, error) {
	switch config.Provider {
	case ProviderEnv:
		return NewEnvSecretManager(config.EnvPrefix), nil
		
	case ProviderAWS:
		if config.AWSRegion == "" || config.AWSSecretArn == "" {
			return nil, fmt.Errorf("AWS region and secret ARN are required for AWS provider")
		}
		return NewAWSSecretManager(ctx, config.AWSRegion, config.AWSSecretArn)
		
	case ProviderGoogle:
		if config.GoogleProjectID == "" || config.GoogleSecretID == "" {
			return nil, fmt.Errorf("Google project ID and secret ID are required for Google provider")
		}
		return NewGoogleSecretManager(ctx, config.GoogleProjectID, config.GoogleSecretID)
		
	default:
		return nil, fmt.Errorf("unsupported secret provider: %s", config.Provider)
	}
}

// DetectProvider automatically detects the secret provider based on environment
func DetectProvider() SecretProvider {
	// Check for Google Cloud environment
	if os.Getenv("GOOGLE_CLOUD_PROJECT") != "" || os.Getenv("GCLOUD_PROJECT") != "" {
		return ProviderGoogle
	}
	
	// Check for AWS environment
	if os.Getenv("AWS_REGION") != "" || os.Getenv("AWS_DEFAULT_REGION") != "" {
		return ProviderAWS
	}
	
	// Default to environment variables
	return ProviderEnv
}

// NewSecretManagerFromEnv creates a secret manager using environment variables for configuration
func NewSecretManagerFromEnv(ctx context.Context) (SecretManager, error) {
	provider := SecretProvider(os.Getenv("SECRET_PROVIDER"))
	if provider == "" {
		provider = DetectProvider()
	}

	config := Config{
		Provider:        provider,
		AWSRegion:       os.Getenv("AWS_REGION"),
		AWSSecretArn:    os.Getenv("AWS_SECRET_ARN"),
		GoogleProjectID: os.Getenv("GOOGLE_PROJECT_ID"),
		GoogleSecretID:  os.Getenv("GOOGLE_SECRET_ID"),
		EnvPrefix:       os.Getenv("SECRET_ENV_PREFIX"),
	}

	return NewSecretManager(ctx, config)
}