package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration.
type Config struct {
	ClerkSecretKey     string
	ClerkWebhookSecret string
	DSN                string
	GRPCServerPort     string
	WebhookServerPort  string
}

// Load loads the configuration from environment variables.
func Load() (*Config, error) {
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY environment variable not set")
	}

	webhookSecret := os.Getenv("CLERK_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, fmt.Errorf("CLERK_WEBHOOK_SECRET environment variable not set")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=user password=password dbname=user_db port=5432 sslmode=disable"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	webhookPort := os.Getenv("WEBHOOK_PORT")
	if webhookPort == "" {
		webhookPort = "8081"
	}

	return &Config{
		ClerkSecretKey:     secretKey,
		ClerkWebhookSecret: webhookSecret,
		DSN:                dsn,
		GRPCServerPort:     grpcPort,
		WebhookServerPort:  webhookPort,
	}, nil
}
