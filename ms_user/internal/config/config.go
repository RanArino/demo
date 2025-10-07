package config

import (
	"context"
	"demo/ms_user/internal/secrets"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ClerkSecretKey        string
	ClerkWebhookSecret    string
	DSN                   string
	GRPCServerPort        string
	WebhookServerPort     string
	WebhookMaxBodySize    int64
	DefaultStorageQuotaGB int64
}

var SecretKeys = []string{
	"CLERK_SECRET_KEY",
	"CLERK_WEBHOOK_SECRET",
	"DATABASE_URL",
}

func Load() (*Config, error) {
	return LoadWithContext(context.Background())
}

func LoadWithContext(ctx context.Context) (*Config, error) {
	if cfg, err := loadFromSecretManager(ctx); err == nil {
		return cfg, nil
	}
	return loadFromEnv()
}

func loadFromSecretManager(ctx context.Context) (*Config, error) {
	secretManager, err := secrets.NewSecretManagerFromEnv(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager: %w", err)
	}
	defer secretManager.Close()

	secretValues, err := secretManager.GetSecrets(ctx, SecretKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	for _, key := range SecretKeys[:1] {
		if secretValues[key] == "" {
			return nil, fmt.Errorf("required secret %s is empty", key)
		}
	}

	dsn, err := ensureSimpleProtocol(secretValues["DATABASE_URL"])
	if err != nil {
		return nil, fmt.Errorf("failed to prepare DSN: %w", err)
	}

	config := &Config{
		ClerkSecretKey:        secretValues["CLERK_SECRET_KEY"],
		ClerkWebhookSecret:    secretValues["CLERK_WEBHOOK_SECRET"],
		DSN:                   dsn,
		GRPCServerPort:        getEnvOrDefault("GRPC_PORT", "50051"),
		WebhookServerPort:     getEnvOrDefault("WEBHOOK_PORT", "8081"),
		WebhookMaxBodySize:    getEnvOrDefaultInt64("WEBHOOK_MAX_BODY_SIZE_MB", 1) << 20,
		DefaultStorageQuotaGB: getEnvOrDefaultInt64("DEFAULT_STORAGE_QUOTA_GB", 5),
	}

	return config, nil
}

func loadFromEnv() (*Config, error) {
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY not set")
	}

	webhookSecret := os.Getenv("CLERK_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, fmt.Errorf("CLERK_WEBHOOK_SECRET not set")
	}

	rawDSN := os.Getenv("DATABASE_URL")
	if rawDSN == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	dsn, err := ensureSimpleProtocol(rawDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare DSN: %w", err)
	}

	return &Config{
		ClerkSecretKey:        secretKey,
		ClerkWebhookSecret:    webhookSecret,
		DSN:                   dsn,
		GRPCServerPort:        getEnvOrDefault("GRPC_PORT", "50051"),
		WebhookServerPort:     getEnvOrDefault("WEBHOOK_PORT", "8081"),
		WebhookMaxBodySize:    getEnvOrDefaultInt64("WEBHOOK_MAX_BODY_SIZE_MB", 1) << 20,
		DefaultStorageQuotaGB: getEnvOrDefaultInt64("DEFAULT_STORAGE_QUOTA_GB", 5),
	}, nil
}

func LoadFromFile(filename string) (*Config, error) {
	if err := loadEnvFile(filename); err != nil {
		return nil, fmt.Errorf("failed to load env file: %w", err)
	}
	return loadFromEnv()
}

func LoadForDevelopment() (*Config, error) {
	if err := loadEnvFile(".env.local"); err == nil {
		return loadFromEnv()
	}
	if err := loadEnvFile(".env"); err == nil {
		return loadFromEnv()
	}
	return Load()
}

func IsProduction() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "production" || env == "prod"
}

func IsDevelopment() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "development" || env == "dev" || env == ""
}

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

		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
			(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			value = value[1 : len(value)-1]
		}

		os.Setenv(key, value)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt64(key string, defaultValue int64) int64 {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetPostgresDSN returns the PostgreSQL DSN string with error handling
func (c *Config) GetPostgresDSN() (string, error) {
	if c.DSN == "" {
		return "", fmt.Errorf("DSN is empty")
	}
	return c.DSN, nil
}

func ensureSimpleProtocol(dsn string) (string, error) {
	if dsn == "" {
		return "", fmt.Errorf("DSN is empty")
	}

	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("failed to parse DSN URL: %w", err)
	}

	q := parsedURL.Query()
	q.Set("prefer_simple_protocol", "1")
	parsedURL.RawQuery = q.Encode()

	return parsedURL.String(), nil
}
