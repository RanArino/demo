package main

import (
	"context"
	"fmt"
	"log"
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
	kmw "demo/ms_knowledge/internal/middleware"
	"demo/ms_knowledge/internal/repository"
	"demo/ms_knowledge/internal/server"
	"demo/ms_knowledge/internal/service"
	storager2 "demo/ms_knowledge/internal/storage/r2"

	userv1 "demo/ms_user/api/proto/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// Initialize repositories
	spaceRepo := repository.NewSpaceRepository(client)
	contentRepo := repository.NewContentRepository(client)

	// Initialize services
	contentLogger := slog.NewLogLogger(handler, slog.LevelInfo)
	spaceService := service.NewSpaceService(spaceRepo, contentRepo)

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
	contentService := service.NewContentService(contentRepo, spaceRepo, r2Storage, producer, cfg, contentLogger)
	// Initialize gRPC server
	grpcServer := server.NewGRPCServer(spaceService, contentService)

	// Initialize User service client for authentication
	var userClient userv1.UserServiceClient
	var userConn *grpc.ClientConn
	userSvcAddr := cfg.Services.UserGRPCAddr
	if userSvcAddr != "" {
		conn, err := grpc.NewClient(userSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			slog.Error("Failed to connect to User service", "address", userSvcAddr, "error", err)
			os.Exit(1)
		}
		userClient = userv1.NewUserServiceClient(conn)
		userConn = conn
		slog.Info("Connected to User service", "address", userSvcAddr)
	} else {
		slog.Error("User service address not configured", "config_key", "Services.UserGRPCAddr")
		os.Exit(1)
	}

	// Wire user service client for server-side identity resolution
	grpcServer = grpcServer.WithUserClient(userClient)

	// Create gRPC server with authentication
	clerkKey := cfg.Auth.ClerkSecretKey
	var serverOpts []grpc.ServerOption
	if clerkKey != "" {
		// Create auth interceptor with User service client dependency
		authI := kmw.NewAuthInterceptor(clerkKey, userClient)
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(authI.Unary()))
		slog.Info("Authentication interceptor enabled", "clerk_configured", true, "user_service_configured", true)
	} else {
		slog.Error("Clerk secret key not configured", "config_key", "Auth.ClerkSecretKey")
		os.Exit(1)
	}
	srv := grpc.NewServer(serverOpts...)
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

	// Close User service connection
	if userConn != nil {
		if err := userConn.Close(); err != nil {
			slog.Warn("Failed to close User service connection", "error", err)
		}
	}

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

	ownerID, err := h.svc.GetContentOwner(ctx, event.ContentSourceID)
	if err != nil {
		return fmt.Errorf("failed to resolve content owner: %w", err)
	}
	ctxWithOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())

	_, err = h.svc.UpdateContentSourceStatus(ctxWithOwner, event.ContentSourceID, status, processedHash, event.ErrorMessage, event.SummaryPtr(), event.KeywordsPtr(), event.TitlePtr())
	return err
}
