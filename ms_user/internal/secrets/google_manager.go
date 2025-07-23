package secrets

import (
	"context"
	"encoding/json"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// GoogleSecretManager retrieves secrets from Google Secret Manager
type GoogleSecretManager struct {
	client    *secretmanager.Client
	projectID string
	secretID  string
	cache     map[string]string
}

// NewGoogleSecretManager creates a new Google Secret Manager client
func NewGoogleSecretManager(ctx context.Context, projectID, secretID string) (*GoogleSecretManager, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	manager := &GoogleSecretManager{
		client:    client,
		projectID: projectID,
		secretID:  secretID,
		cache:     make(map[string]string),
	}

	// Load secrets into cache on initialization
	if err := manager.RefreshSecrets(ctx); err != nil {
		return nil, fmt.Errorf("failed to load initial secrets: %w", err)
	}

	return manager, nil
}

// GetSecret retrieves a specific secret
func (g *GoogleSecretManager) GetSecret(ctx context.Context, key string) (string, error) {
	if value, exists := g.cache[key]; exists {
		return value, nil
	}
	
	// Try to refresh cache and check again
	if err := g.RefreshSecrets(ctx); err != nil {
		return "", fmt.Errorf("failed to refresh secrets: %w", err)
	}
	
	if value, exists := g.cache[key]; exists {
		return value, nil
	}
	
	return "", fmt.Errorf("secret key %s not found", key)
}

// GetSecrets retrieves multiple secrets
func (g *GoogleSecretManager) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	secrets := make(map[string]string)
	var missingKeys []string

	for _, key := range keys {
		if value, exists := g.cache[key]; exists {
			secrets[key] = value
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	// If we have missing keys, try refreshing the cache
	if len(missingKeys) > 0 {
		if err := g.RefreshSecrets(ctx); err != nil {
			return secrets, fmt.Errorf("failed to refresh secrets: %w", err)
		}

		// Check for missing keys again after refresh
		var stillMissing []string
		for _, key := range missingKeys {
			if value, exists := g.cache[key]; exists {
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

// RefreshSecrets refreshes the secret cache from Google Secret Manager
func (g *GoogleSecretManager) RefreshSecrets(ctx context.Context) error {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", g.projectID, g.secretID),
	}

	result, err := g.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to access secret version: %w", err)
	}

	// Parse JSON secrets
	var secrets map[string]string
	if err := json.Unmarshal(result.Payload.Data, &secrets); err != nil {
		return fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	// Update cache
	g.cache = secrets
	return nil
}

// Close closes the Google client
func (g *GoogleSecretManager) Close() error {
	return g.client.Close()
}