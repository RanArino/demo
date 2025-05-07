package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/ran/demo/backend-go/internal/api"
	appconfig "github.com/ran/demo/backend-go/internal/config"
	"github.com/ran/demo/backend-go/internal/core/models/document"
	documentService "github.com/ran/demo/backend-go/internal/core/services/document"
	parserService "github.com/ran/demo/backend-go/internal/core/services/document/parser"
	parserStrategies "github.com/ran/demo/backend-go/internal/core/services/document/parser/strategies"
	sqsQueue "github.com/ran/demo/backend-go/internal/infra/queue/sqs"
	"github.com/ran/demo/backend-go/internal/infra/storage/postgres"
	s3Storage "github.com/ran/demo/backend-go/internal/infra/storage/s3"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg, err := appconfig.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := setupDatabase(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Setup AWS services
	awsCfg, err := appconfig.LoadAWSConfig(cfg)
	if err != nil {
		logger.Error("Failed to load AWS configuration", "error", err)
		os.Exit(1)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.S3.Endpoint != "" {
			o.BaseEndpoint = &cfg.S3.Endpoint
		}
		o.UsePathStyle = cfg.S3.ForcePathStyle
	})
	_ = s3Client // Using s3Client via s3StorageClient below

	// Create SQS client
	sqsClient := sqs.NewFromConfig(awsCfg)

	// Create S3 storage service
	s3Config := &document.S3Config{
		Region:               cfg.S3.Region,
		Endpoint:             cfg.S3.Endpoint,
		Bucket:               cfg.S3.Bucket,
		ForcePathStyle:       cfg.S3.ForcePathStyle,
		DisableSSL:           cfg.S3.DisableSSL,
		PresignedURLDuration: cfg.S3.PresignedURLDuration,
	}
	s3StorageClient, err := s3Storage.NewS3Client(s3Config, logger)
	if err != nil {
		logger.Error("Failed to create S3 client", "error", err)
		os.Exit(1)
	}
	storageService := s3Storage.NewS3StorageService(s3StorageClient, logger)

	// Create SQS queue service
	queueService := sqsQueue.NewSQSQueueService(
		sqsClient,
		cfg.SQS.QueueURL,
		cfg.SQS.DeadLetterQueueURL,
		logger,
	)

	// Create document repository
	documentRepo := postgres.NewDocumentRepository(db, logger)

	// Create parser orchestrator
	parserOrchestrator := parserService.NewFileParsingOrchestrator(
		storageService,
		nil, // No converter service yet
		logger,
	)

	// Register parser strategies
	markdownStrategy := parserStrategies.NewMarkdownStrategy(storageService, logger)
	err = parserOrchestrator.RegisterStrategy(markdownStrategy)
	if err != nil {
		logger.Error("Failed to register markdown strategy", "error", err)
		os.Exit(1)
	}

	// Create document presenter
	documentPresenter := documentService.NewDocumentPresenter(
		documentRepo,
		storageService,
		logger,
	)

	// Create document lifecycle manager
	documentLifecycleManager := documentService.NewDocumentLifecycleManager(
		documentRepo,
		storageService,
		queueService,
		parserOrchestrator,
		documentPresenter,
		logger,
	)

	// Create API server
	server := api.NewServer(
		cfg,
		documentLifecycleManager,
		documentPresenter,
		logger,
	)

	// Start server
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown gracefully
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
	}

	logger.Info("Server stopped")
}

// setupDatabase connects to the PostgreSQL database
func setupDatabase(cfg *appconfig.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeSeconds) * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
