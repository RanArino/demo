package config

import (
	"context"
	"demo/ms_user/internal/secrets"
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	ClerkSecretKey    string
	ClerkWebhookSecret string
	DSN               string
	GRPCServerPort    string
	WebhookServerPort string
}

// SecretKeys defines the keys needed from the secret manager
var SecretKeys = []string{
	"CLERK_SECRET_KEY",
	"CLERK_WEBHOOK_SECRET",
	"DATABASE_URL",
}

// Load loads the configuration from the secret manager with fallback to environment variables.
func Load() (*Config, error) {
	return LoadWithContext(context.Background())
}

// LoadWithContext loads the configuration with a specific context
func LoadWithContext(ctx context.Context) (*Config, error) {
	// Try to load from secret manager first
	if cfg, err := loadFromSecretManager(ctx); err == nil {
		return cfg, nil
	}

	// Fallback to environment variables
	return loadFromEnv()
}

// loadFromSecretManager loads configuration from cloud secret managers
func loadFromSecretManager(ctx context.Context) (*Config, error) {
	secretManager, err := secrets.NewSecretManagerFromEnv(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager: %w", err)
	}
	defer secretManager.Close()

	// Get required secrets
	secretValues, err := secretManager.GetSecrets(ctx, SecretKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	// Validate required secrets
	for _, key := range SecretKeys[:1] { // Only CLERK_SECRET_KEY is required
		if secretValues[key] == "" {
			return nil, fmt.Errorf("required secret %s is empty", key)
		}
	}

	config := &Config{
		ClerkSecretKey:     secretValues["CLERK_SECRET_KEY"],
		ClerkWebhookSecret: secretValues["CLERK_WEBHOOK_SECRET"],
		DSN:                secretValues["DATABASE_URL"],
		GRPCServerPort:     getEnvOrDefault("GRPC_PORT", "50051"),
		WebhookServerPort:  getEnvOrDefault("WEBHOOK_PORT", "8081"),
	}

	return config, nil
}

// loadFromEnv loads configuration from environment variables (fallback)
func loadFromEnv() (*Config, error) {
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY not set")
	}

	webhookSecret := os.Getenv("CLERK_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, fmt.Errorf("CLERK_WEBHOOK_SECRET not set")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	return &Config{
		ClerkSecretKey:    secretKey,
		ClerkWebhookSecret: webhookSecret,
		DSN:               dsn,
		GRPCServerPort:    getEnvOrDefault("GRPC_PORT", "50051"),
		WebhookServerPort: getEnvOrDefault("WEBHOOK_PORT", "8081"),
	}, nil
}

// LoadFromFile loads configuration from a .env file (for development)
func LoadFromFile(filename string) (*Config, error) {
	if err := loadEnvFile(filename); err != nil {
		return nil, fmt.Errorf("failed to load env file: %w", err)
	}
	return loadFromEnv()
}

// LoadForDevelopment loads configuration with development-specific logic
func LoadForDevelopment() (*Config, error) {
	// Try to load from .env.local first
	if err := loadEnvFile(".env.local"); err == nil {
		return loadFromEnv()
	}

	// Try to load from .env
	if err := loadEnvFile(".env"); err == nil {
		return loadFromEnv()
	}

	// Fallback to regular loading process
	return Load()
}

// IsProduction determines if the application is running in production
func IsProduction() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "production" || env == "prod"
}

// IsDevelopment determines if the application is running in development
func IsDevelopment() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "development" || env == "dev" || env == ""
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
			(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			value = value[1 : len(value)-1]
		}

		os.Setenv(key, value)
	}

	return nil
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
