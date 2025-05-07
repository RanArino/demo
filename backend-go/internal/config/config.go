package config

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
)

// Config contains application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	S3       S3Config
	SQS      SQSConfig
}

// ServerConfig contains web server configuration
type ServerConfig struct {
	Port            int
	Host            string
	ReadTimeout     int
	WriteTimeout    int
	ShutdownTimeout int
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	URL                    string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeSeconds int
}

// S3Config contains S3 configuration
type S3Config struct {
	Region                string
	Bucket                string
	Endpoint              string
	AccessKey             string
	SecretKey             string
	UseIAMRoleCredentials bool
	ForcePathStyle        bool
	DisableSSL            bool
	PresignedURLDuration  int
}

// SQSConfig contains SQS configuration
type SQSConfig struct {
	QueueURL           string
	DeadLetterQueueURL string
	Region             string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Get SQS region, defaulting to S3 region if not set
	sqsRegion := getEnv("SQS_REGION", getEnv("AWS_REGION", "us-east-1"))

	return &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			Host:            getEnv("SERVER_HOST", ""),
			ReadTimeout:     getEnvAsInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout:    getEnvAsInt("SERVER_WRITE_TIMEOUT", 10),
			ShutdownTimeout: getEnvAsInt("SERVER_SHUTDOWN_TIMEOUT", 5),
		},
		Database: DatabaseConfig{
			URL:                    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/document_service?sslmode=disable"),
			MaxOpenConns:           getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:           getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 25),
			ConnMaxLifetimeSeconds: getEnvAsInt("DATABASE_CONN_MAX_LIFETIME_SECONDS", 300),
		},
		S3: S3Config{
			Region:                getEnv("AWS_REGION", "us-east-1"),
			Bucket:                getEnv("S3_BUCKET", ""),   // Expect this from Terraform outputs
			Endpoint:              getEnv("S3_ENDPOINT", ""), // For LocalStack/MinIO
			AccessKey:             getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretKey:             getEnv("AWS_SECRET_ACCESS_KEY", ""),
			UseIAMRoleCredentials: getEnvAsBool("AWS_USE_IAM_ROLE", false),
			ForcePathStyle:        getEnvAsBool("S3_FORCE_PATH_STYLE", false),
			DisableSSL:            getEnvAsBool("S3_DISABLE_SSL", false),
			PresignedURLDuration:  getEnvAsInt("S3_PRESIGNED_URL_DURATION", 900),
		},
		SQS: SQSConfig{
			QueueURL:           getEnv("SQS_QUEUE_URL", ""),             // Expect this from Terraform outputs
			DeadLetterQueueURL: getEnv("SQS_DEAD_LETTER_QUEUE_URL", ""), // Expect this from Terraform outputs
			Region:             sqsRegion,
		},
	}, nil
}

// LoadAWSConfig loads AWS configuration for services
func LoadAWSConfig(cfg *Config) (aws.Config, error) {
	// Use AWS SDK's default credential chain, which includes:
	// 1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
	// 2. Shared credentials file (~/.aws/credentials)
	// 3. IAM role for Amazon EC2/ECS
	// 4. Container credentials (ECS task IAM roles)
	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.S3.Region),
	}

	// If we have a role ARN to assume, configure that
	if roleARN := getEnv("AWS_ROLE_ARN", ""); roleARN != "" {
		optFns = append(optFns, awsconfig.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
			options.RoleARN = roleARN
			// External ID is optional, so only set if provided
			externalID := getEnv("AWS_EXTERNAL_ID", "")
			if externalID != "" {
				options.ExternalID = aws.String(externalID)
			}
		}))
	} else if !cfg.S3.UseIAMRoleCredentials && cfg.S3.AccessKey != "" && cfg.S3.SecretKey != "" {
		// If not using instance role and have explicit credentials
		optFns = append(optFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, ""),
		))
	}

	// For LocalStack/MinIO
	if cfg.S3.Endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.S3.Endpoint,
				HostnameImmutable: true,
				SigningRegion:     cfg.S3.Region,
			}, nil
		})
		optFns = append(optFns, awsconfig.WithEndpointResolverWithOptions(customResolver))
	}

	// For LocalStack - skip validation checks
	if cfg.S3.Endpoint != "" && strings.Contains(cfg.S3.Endpoint, "localhost") {
		optFns = append(optFns,
			awsconfig.WithClientLogMode(aws.LogSigning|aws.LogRequest|aws.LogResponse),
			awsconfig.WithRetryMaxAttempts(1),
			// Use anonymous credentials for local development
			func(options *awsconfig.LoadOptions) error {
				options.Credentials = aws.AnonymousCredentials{}
				return nil
			},
		)
	}

	// Load AWS configuration
	ctx := context.Background()
	return awsconfig.LoadDefaultConfig(ctx, optFns...)
}

// Helper functions to get environment variables
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsStringSlice(key string, defaultValue []string, separator string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, separator)
}
