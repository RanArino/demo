package config

import (
	"context"
	"demo/ms_knowledge/internal/secrets"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	DatabaseURL       string
	GRPCPort          string
	Neo4jURI          string
	Neo4jUser         string
	Neo4jPassword     string
	KafkaBrokers      string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2AccountID       string
	R2BucketName      string
}

// SecretKeys defines the keys needed from the secret manager
var SecretKeys = []string{
	"DATABASE_URL",
	"NEO4J_URI",
	"NEO4J_USER",
	"NEO4J_PASSWORD",
	"KAFKA_BROKERS",
	"R2_ACCESS_KEY_ID",
	"R2_SECRET_ACCESS_KEY",
	"R2_ACCOUNT_ID",
	"R2_BUCKET_NAME",
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

	config := &Config{
		DatabaseURL:       secretValues["DATABASE_URL"],
		Neo4jURI:          secretValues["NEO4J_URI"],
		Neo4jUser:         secretValues["NEO4J_USER"],
		Neo4jPassword:     secretValues["NEO4J_PASSWORD"],
		KafkaBrokers:      secretValues["KAFKA_BROKERS"],
		R2AccessKeyID:     secretValues["R2_ACCESS_KEY_ID"],
		R2SecretAccessKey: secretValues["R2_SECRET_ACCESS_KEY"],
		R2AccountID:       secretValues["R2_ACCOUNT_ID"],
		R2BucketName:      secretValues["R2_BUCKET_NAME"],
		GRPCPort:          getEnvOrDefault("GRPC_PORT", "50052"),
	}

	return config, nil
}

// loadFromEnv loads configuration from environment variables (fallback)
func loadFromEnv() (*Config, error) {
	return &Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		Neo4jURI:          os.Getenv("NEO4J_URI"),
		Neo4jUser:         os.Getenv("NEO4J_USER"),
		Neo4jPassword:     os.Getenv("NEO4J_PASSWORD"),
		KafkaBrokers:      os.Getenv("KAFKA_BROKERS"),
		R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		R2BucketName:      os.Getenv("R2_BUCKET_NAME"),
		GRPCPort:          getEnvOrDefault("GRPC_PORT", "50052"),
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
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
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

func getEnvOrDefaultInt64(key string, defaultValue int64) int64 {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			return value
		}
	}
	return defaultValue
}
