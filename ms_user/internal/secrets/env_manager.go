package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EnvSecretManager retrieves secrets from environment variables
type EnvSecretManager struct {
	prefix string
}

// NewEnvSecretManager creates a new environment-based secret manager
func NewEnvSecretManager(prefix string) *EnvSecretManager {
	return &EnvSecretManager{
		prefix: prefix,
	}
}

// GetSecret retrieves a secret from environment variables
func (e *EnvSecretManager) GetSecret(ctx context.Context, key string) (string, error) {
	envKey := e.formatKey(key)
	value := os.Getenv(envKey)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not set", envKey)
	}
	return value, nil
}

// GetSecrets retrieves multiple secrets from environment variables
func (e *EnvSecretManager) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	secrets := make(map[string]string)
	var missingKeys []string

	for _, key := range keys {
		envKey := e.formatKey(key)
		if value := os.Getenv(envKey); value != "" {
			secrets[key] = value
		} else {
			missingKeys = append(missingKeys, envKey)
		}
	}

	if len(missingKeys) > 0 {
		return secrets, fmt.Errorf("environment variables not set: %s", strings.Join(missingKeys, ", "))
	}

	return secrets, nil
}

// RefreshSecrets is a no-op for environment variables
func (e *EnvSecretManager) RefreshSecrets(ctx context.Context) error {
	return nil
}

// Close is a no-op for environment variables
func (e *EnvSecretManager) Close() error {
	return nil
}

// formatKey formats the key with prefix if provided
func (e *EnvSecretManager) formatKey(key string) string {
	if e.prefix != "" {
		return e.prefix + "_" + key
	}
	return key
}