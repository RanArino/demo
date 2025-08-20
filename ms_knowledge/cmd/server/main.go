package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	knowledgev1 "demo/ms_knowledge/api/proto/v1"
	"demo/ms_knowledge/ent"
	"demo/ms_knowledge/internal/config"
	"demo/ms_knowledge/internal/domain"
	"demo/ms_knowledge/internal/events"
	"demo/ms_knowledge/internal/repository"
	"demo/ms_knowledge/internal/repository/graph"
	"demo/ms_knowledge/internal/server"
	"demo/ms_knowledge/internal/service"
	storager2 "demo/ms_knowledge/internal/storage/r2"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize structured logger
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Create a context that can be cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection
	client, err := ent.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// Run database migrations
	if err := client.Schema.Create(ctx); err != nil {
		slog.Error("Failed to run database migrations", "error", err)
		os.Exit(1)
	}

	// Initialize Neo4j connection
	neo4jDriver, err := neo4j.NewDriverWithContext(
		cfg.Neo4j.URI,
		neo4j.BasicAuth(cfg.Neo4j.Username, cfg.Neo4j.Password, ""),
	)
	if err != nil {
		slog.Error("Failed to connect to Neo4j", "error", err)
		os.Exit(1)
	}
	defer neo4jDriver.Close(ctx)

	// Verify Neo4j connection
	if err := neo4jDriver.VerifyConnectivity(ctx); err != nil {
		slog.Error("Failed to verify Neo4j connectivity", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	graphRepo := graph.NewNeo4jRepository(neo4jDriver)
	spaceRepo := repository.NewSpaceRepository(client, graphRepo)
	contentRepo := repository.NewContentRepository(client, graphRepo)

	// Initialize services
	contentLogger := slog.NewLogLogger(handler, slog.LevelInfo)
	spaceService := service.NewSpaceService(spaceRepo, contentRepo, graphRepo)

	// Initialize R2 storage client
	r2Client, err := storager2.NewClient(ctx, storager2.Config{
		Endpoint:        cfg.R2.Endpoint,
		Region:          cfg.R2.Region,
		AccessKeyID:     cfg.R2.AccessKeyID,
		SecretAccessKey: cfg.R2.SecretAccessKey,
		UsePathStyle:    true,
	})
	if err != nil {
		slog.Error("Failed to create R2 client", "error", err)
		os.Exit(1)
	}

	// Initialize Kafka producer and consumer for events
	producer, err := events.NewProducer(cfg)
	if err != nil {
		slog.Error("Failed to create Kafka producer", "error", err)
		os.Exit(1)
	}
	defer producer.Close()

	r2Storage := storager2.NewAdapter(r2Client)
	contentService := service.NewContentService(contentRepo, spaceRepo, graphRepo, r2Storage, producer, cfg.R2.BucketSourceName, contentLogger)
	knowledgeLinkService := service.NewKnowledgeLinkService(graphRepo, contentRepo)

	// Initialize gRPC server
	grpcServer := server.NewGRPCServer(spaceService, contentService, knowledgeLinkService)

	// Create gRPC server
	srv := grpc.NewServer()
	knowledgev1.RegisterKnowledgeServiceServer(srv, grpcServer)

	// Enable reflection for development
	reflection.Register(srv)

	// Start document.processed consumer in background
	consumer, err := events.NewConsumer(cfg, "ms_knowledge-processed-group", &processedHandler{svc: contentService})
	if err != nil {
		slog.Error("Failed to create Kafka consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()
	go consumer.Run(ctx)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		slog.Error("Failed to listen", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting gRPC server", "port", cfg.Server.Port)
	go func() {
		if err := srv.Serve(lis); err != nil {
			slog.Error("Failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown
	slog.Info("Shutting down server...")
	cancel() // Cancel context for consumer
	srv.GracefulStop()
	slog.Info("Server stopped")
}

// mapProcessStatusToContentStatus safely maps events.ProcessStatus to domain.ContentStatus.
func mapProcessStatusToContentStatus(status events.ProcessStatus) (domain.ContentStatus, error) {
	switch status {
	case events.ProcessStatusPending:
		return domain.ContentStatusPending, nil
	case events.ProcessStatusProcessing:
		return domain.ContentStatusProcessing, nil
	case events.ProcessStatusProcessed:
		return domain.ContentStatusProcessed, nil
	case events.ProcessStatusFailed:
		return domain.ContentStatusFailed, nil
	default:
		return "", fmt.Errorf("unknown ProcessStatus: %v", status)
	}
}

// processedHandler adapts the consumer callback to the service method.
type processedHandler struct {
	svc *service.ContentService
}

func (h *processedHandler) HandleDocumentProcessed(ctx context.Context, event events.DocumentProcessedEvent) error {
	status, err := mapProcessStatusToContentStatus(event.Status)
	if err != nil {
		log.Printf("Invalid ProcessStatus in event, status %v, error %v", event.Status, err)
		return err
	}
	processedHash := ""
	if event.ProcessedBlobHash != nil {
		processedHash = *event.ProcessedBlobHash
	}
	_, err = h.svc.UpdateContentSourceStatus(ctx, event.ContentSourceID, status, processedHash, event.ErrorMessage)
	return err
}
