package config

import (
	"context"
	"demo/ms_knowledge/internal/secrets"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	Database struct {
		Driver string
		DSN    string
	}
	Server struct {
		Port int
	}
	Kafka struct {
		Brokers          string
		SaslUsername     string
		SaslPassword     string
		SecurityProtocol string
	}
	R2 struct {
		AccessKeyID         string
		SecretAccessKey     string
		AccountID           string
		Endpoint            string
		Region              string
		BucketSourceName    string
		BucketProcessedName string
	}
	Auth struct {
		ClerkSecretKey string
	}
	Services struct {
		UserGRPCAddr string
	}
}

// SecretKeys defines the keys needed from the secret manager
var SecretKeys = []string{
	"DATABASE_URL",
	"KAFKA_BROKERS",
	"KAFKA_SASL_USERNAME",
	"KAFKA_SASL_PASSWORD",
	"KAFKA_SECURITY_PROTOCOL",
	"R2_ACCESS_KEY_ID",
	"R2_SECRET_ACCESS_KEY",
	"R2_ACCOUNT_ID",
	"R2_ENDPOINT",
	"R2_REGION",
	"R2_BUCKET_SOURCE_NAME",
	"R2_BUCKET_PROCESSED_NAME",
	"CLERK_SECRET_KEY",
	"MS_USER_GRPC_URL_INTERNAL",
}

// Load loads the configuration from the secret manager with fallback to environment variables.
func Load() (*Config, error) {
	cfg := &Config{}

	// Database configuration
	cfg.Database.Driver = "postgres"
	cfg.Database.DSN = getEnvOrDefault("DATABASE_URL", "postgres://user:password@localhost:5432/knowledge?sslmode=disable")

	// Server configuration
	cfg.Server.Port = int(getEnvOrDefaultInt64("GRPC_PORT", 50052))


	// Kafka configuration
	cfg.Kafka.Brokers = getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")
	cfg.Kafka.SaslUsername = getEnvOrDefault("KAFKA_SASL_USERNAME", "")
	cfg.Kafka.SaslPassword = getEnvOrDefault("KAFKA_SASL_PASSWORD", "")
	cfg.Kafka.SecurityProtocol = getEnvOrDefault("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT")

	// R2 configuration
	cfg.R2.AccessKeyID = getEnvOrDefault("R2_ACCESS_KEY_ID", "")
	cfg.R2.SecretAccessKey = getEnvOrDefault("R2_SECRET_ACCESS_KEY", "")
	cfg.R2.AccountID = getEnvOrDefault("R2_ACCOUNT_ID", "")
	cfg.R2.Endpoint = getEnvOrDefault("R2_ENDPOINT", "")
	cfg.R2.Region = getEnvOrDefault("R2_REGION", "auto")
	cfg.R2.BucketSourceName = getEnvOrDefault("R2_BUCKET_SOURCE_NAME", "knowledge-source")
	cfg.R2.BucketProcessedName = getEnvOrDefault("R2_BUCKET_PROCESSED_NAME", "knowledge-processed")

	// Auth configuration
	cfg.Auth.ClerkSecretKey = getEnvOrDefault("CLERK_SECRET_KEY", "")

	// Services configuration
	cfg.Services.UserGRPCAddr = getEnvOrDefault("MS_USER_GRPC_URL_INTERNAL", "localhost:50051")

	return cfg, nil
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

	config := &Config{}

	// Database configuration
	config.Database.Driver = "postgres"
	config.Database.DSN = secretValues["DATABASE_URL"]

	// Server configuration
	config.Server.Port = int(getEnvOrDefaultInt64("GRPC_PORT", 50052))


	// Kafka configuration
	config.Kafka.Brokers = secretValues["KAFKA_BROKERS"]
	config.Kafka.SaslUsername = secretValues["KAFKA_SASL_USERNAME"]
	config.Kafka.SaslPassword = secretValues["KAFKA_SASL_PASSWORD"]
	config.Kafka.SecurityProtocol = secretValues["KAFKA_SECURITY_PROTOCOL"]

	// R2 configuration
	config.R2.AccessKeyID = secretValues["R2_ACCESS_KEY_ID"]
	config.R2.SecretAccessKey = secretValues["R2_SECRET_ACCESS_KEY"]
	config.R2.AccountID = secretValues["R2_ACCOUNT_ID"]
	config.R2.Endpoint = secretValues["R2_ENDPOINT"]
	config.R2.Region = secretValues["R2_REGION"]
	config.R2.BucketSourceName = secretValues["R2_BUCKET_SOURCE_NAME"]
	config.R2.BucketProcessedName = secretValues["R2_BUCKET_PROCESSED_NAME"]

	// Auth configuration
	config.Auth.ClerkSecretKey = secretValues["CLERK_SECRET_KEY"]

	// Services configuration
	config.Services.UserGRPCAddr = secretValues["MS_USER_GRPC_URL_INTERNAL"]

	return config, nil
}

// loadFromEnv loads configuration from environment variables (fallback)
func loadFromEnv() (*Config, error) {
	config := &Config{}

	// Database configuration
	config.Database.Driver = "postgres"
	config.Database.DSN = os.Getenv("DATABASE_URL")

	// Server configuration
	config.Server.Port = int(getEnvOrDefaultInt64("GRPC_PORT", 50052))


	// Kafka configuration
	config.Kafka.Brokers = os.Getenv("KAFKA_BROKERS")
	config.Kafka.SaslUsername = os.Getenv("KAFKA_SASL_USERNAME")
	config.Kafka.SaslPassword = os.Getenv("KAFKA_SASL_PASSWORD")
	config.Kafka.SecurityProtocol = os.Getenv("KAFKA_SECURITY_PROTOCOL")

	// R2 configuration
	config.R2.AccessKeyID = os.Getenv("R2_ACCESS_KEY_ID")
	config.R2.SecretAccessKey = os.Getenv("R2_SECRET_ACCESS_KEY")
	config.R2.AccountID = os.Getenv("R2_ACCOUNT_ID")
	config.R2.Endpoint = os.Getenv("R2_ENDPOINT")
	config.R2.Region = os.Getenv("R2_REGION")
	config.R2.BucketSourceName = os.Getenv("R2_BUCKET_SOURCE_NAME")
	config.R2.BucketProcessedName = os.Getenv("R2_BUCKET_PROCESSED_NAME")

	// Auth configuration
	config.Auth.ClerkSecretKey = os.Getenv("CLERK_SECRET_KEY")

	// Services configuration
	config.Services.UserGRPCAddr = getEnvOrDefault("MS_USER_GRPC_URL_INTERNAL", "localhost:50051")

	return config, nil
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
	// Resolve filename by searching upwards from CWD so tests in subdirs can find project-root .env files
	resolved := filename
	if _, err := os.Stat(resolved); os.IsNotExist(err) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		for dir := cwd; ; dir = filepath.Dir(dir) {
			candidate := filepath.Join(dir, filename)
			if _, statErr := os.Stat(candidate); statErr == nil {
				resolved = candidate
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir { // reached root
				break
			}
		}
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
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
