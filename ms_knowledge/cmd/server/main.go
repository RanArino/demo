package main

import (
	"context"
	"fmt"
	"log"
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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	client, err := ent.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Close()

	// Run database migrations
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize Neo4j connection
	neo4jDriver, err := neo4j.NewDriverWithContext(
		cfg.Neo4j.URI,
		neo4j.BasicAuth(cfg.Neo4j.Username, cfg.Neo4j.Password, ""),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer neo4jDriver.Close(context.Background())

	// Verify Neo4j connection
	if err := neo4jDriver.VerifyConnectivity(context.Background()); err != nil {
		log.Fatalf("Failed to verify Neo4j connectivity: %v", err)
	}

	// Initialize repositories
	graphRepo := graph.NewNeo4jRepository(neo4jDriver)
	spaceRepo := repository.NewSpaceRepository(client, graphRepo)
	contentRepo := repository.NewContentRepository(client, graphRepo)

	// Initialize services
	logger := log.New(os.Stdout, "", log.LstdFlags)
	spaceService := service.NewSpaceService(spaceRepo, contentRepo, graphRepo)

	// Initialize R2 storage client
	r2Client, err := storager2.NewClient(context.Background(), storager2.Config{
		Endpoint:        cfg.R2.Endpoint,
		Region:          cfg.R2.Region,
		AccessKeyID:     cfg.R2.AccessKeyID,
		SecretAccessKey: cfg.R2.SecretAccessKey,
		UsePathStyle:    true,
	})
	if err != nil {
		log.Fatalf("Failed to create R2 client: %v", err)
	}

	// Initialize Kafka producer and consumer for events
	producer, err := events.NewProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	r2Storage := storager2.NewAdapter(r2Client)
	contentService := service.NewContentService(contentRepo, spaceRepo, graphRepo, r2Storage, producer, cfg.R2.BucketSourceName, logger)
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
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()
	go consumer.Run(context.Background())

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Starting gRPC server on port %d", cfg.Server.Port)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	srv.GracefulStop()
	log.Println("Server stopped")
}

// processedHandler adapts the consumer callback to the service method.
type processedHandler struct {
	svc *service.ContentService
}

func (h *processedHandler) HandleDocumentProcessed(ctx context.Context, event events.DocumentProcessedEvent) error {
	status := domain.ContentStatus(event.Status)
	processedHash := ""
	if event.ProcessedBlobHash != nil {
		processedHash = *event.ProcessedBlobHash
	}
	_, err := h.svc.UpdateContentSourceStatus(ctx, event.ContentSourceID, status, processedHash, event.ErrorMessage)
	return err
}
