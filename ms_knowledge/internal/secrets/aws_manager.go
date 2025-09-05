package secrets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// AWSSecretManager retrieves secrets from AWS Secrets Manager
type AWSSecretManager struct {
	client    *secretsmanager.Client
	secretArn string
	cache     map[string]string
}

// NewAWSSecretManager creates a new AWS Secrets Manager client
func NewAWSSecretManager(ctx context.Context, region, secretArn string) (*AWSSecretManager, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	
	manager := &AWSSecretManager{
		client:    client,
		secretArn: secretArn,
		cache:     make(map[string]string),
	}

	// Load secrets into cache on initialization
	if err := manager.RefreshSecrets(ctx); err != nil {
		return nil, fmt.Errorf("failed to load initial secrets: %w", err)
	}

	return manager, nil
}

// GetSecret retrieves a specific secret
func (a *AWSSecretManager) GetSecret(ctx context.Context, key string) (string, error) {
	if value, exists := a.cache[key]; exists {
		return value, nil
	}
	
	// Try to refresh cache and check again
	if err := a.RefreshSecrets(ctx); err != nil {
		return "", fmt.Errorf("failed to refresh secrets: %w", err)
	}
	
	if value, exists := a.cache[key]; exists {
		return value, nil
	}
	
	return "", fmt.Errorf("secret key %s not found", key)
}

// GetSecrets retrieves multiple secrets
func (a *AWSSecretManager) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	secrets := make(map[string]string)
	var missingKeys []string

	for _, key := range keys {
		if value, exists := a.cache[key]; exists {
			secrets[key] = value
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	// If we have missing keys, try refreshing the cache
	if len(missingKeys) > 0 {
		if err := a.RefreshSecrets(ctx); err != nil {
			return secrets, fmt.Errorf("failed to refresh secrets: %w", err)
		}

		// Check for missing keys again after refresh
		var stillMissing []string
		for _, key := range missingKeys {
			if value, exists := a.cache[key]; exists {
				secrets[key] = value
			} else {
				stillMissing = append(stillMissing, key)
			}
		}

		if len(stillMissing) > 0 {
			return secrets, fmt.Errorf("secrets not found: %v", stillMissing)
		}
	}

	return secrets, nil
}

// RefreshSecrets refreshes the secret cache from AWS
func (a *AWSSecretManager) RefreshSecrets(ctx context.Context) error {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(a.secretArn),
	}

	result, err := a.client.GetSecretValue(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get secret value: %w", err)
	}

	if result.SecretString == nil {
		return fmt.Errorf("secret value is nil")
	}

	// Parse JSON secrets
	var secrets map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
		return fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	// Update cache
	a.cache = secrets
	return nil
}

// Close closes the AWS client (no-op for AWS SDK v2)
func (a *AWSSecretManager) Close() error {
	return nil
}