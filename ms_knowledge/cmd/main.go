package main

import (
	"context"
	"database/sql"
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

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := connectDB(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	if err := client.Schema.Create(ctx); err != nil {
		slog.Error("Failed to run database migrations", "error", err)
		os.Exit(1)
	}

	spaceRepo := repository.NewSpaceRepository(client)
	contentRepo := repository.NewContentRepository(client)

	contentLogger := slog.NewLogLogger(handler, slog.LevelInfo)
	spaceService := service.NewSpaceService(spaceRepo, contentRepo)

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

	producer, err := events.NewProducer(cfg)
	if err != nil {
		slog.Error("Failed to create Kafka producer", "error", err)
		os.Exit(1)
	}
	defer producer.Close()

	r2Storage := storager2.NewAdapter(r2Client)
	contentService := service.NewContentService(contentRepo, spaceRepo, r2Storage, producer, cfg, contentLogger)
	grpcServer := server.NewGRPCServer(spaceService, contentService)

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

	grpcServer = grpcServer.WithUserClient(userClient)

	clerkKey := cfg.Auth.ClerkSecretKey
	var serverOpts []grpc.ServerOption
	if clerkKey != "" {
		authI := kmw.NewAuthInterceptor(clerkKey, userClient)
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(authI.Unary()))
		slog.Info("Authentication interceptor enabled", "clerk_configured", true, "user_service_configured", true)
	} else {
		slog.Error("Clerk secret key not configured", "config_key", "Auth.ClerkSecretKey")
		os.Exit(1)
	}
	srv := grpc.NewServer(serverOpts...)
	knowledgev1.RegisterKnowledgeServiceServer(srv, grpcServer)

	reflection.Register(srv)

	consumer, err := events.NewConsumer(cfg, "ms_knowledge-processed-group", &processedHandler{svc: contentService})
	if err != nil {
		slog.Error("Failed to create Kafka consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()
	go consumer.Run(ctx)

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	cancel()
	srv.GracefulStop()

	if userConn != nil {
		if err := userConn.Close(); err != nil {
			slog.Warn("Failed to close User service connection", "error", err)
		}
	}

	slog.Info("Server stopped")
}

func connectDB(cfg *config.Config) (*sql.DB, error) {
	dsn, err := cfg.GetPostgresDSN()
	if err != nil {
		return nil, fmt.Errorf("failed to get DSN: %w", err)
	}

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	db := stdlib.OpenDB(*pgxConfig)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Successfully connected to database", "host", pgxConfig.Host, "port", pgxConfig.Port)
	return db, nil
}

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

type processedHandler struct {
	svc *service.ContentService
}

func (h *processedHandler) HandleDocumentProcessed(ctx context.Context, event events.DocumentProcessedEvent) error {
	status, err := mapProcessStatusToContentStatus(event.Status)
	if err != nil {
		log.Printf("Invalid ProcessStatus in event, status %v, error %v", event.Status, err)
		return err
	}

	var blobHash string
	if event.ProcessedBlobHash != nil {
		blobHash = *event.ProcessedBlobHash
	}

	ownerID, err := h.svc.GetContentOwner(ctx, event.ContentSourceID)
	if err != nil {
		return fmt.Errorf("failed to resolve content owner: %w", err)
	}
	ctxWithOwner := context.WithValue(ctx, domain.OwnerIDKey, ownerID.String())

	_, err = h.svc.UpdateContentSourceStatus(ctxWithOwner, event.ContentSourceID, status, blobHash, event.ErrorMessage, event.SummaryPtr(), event.KeywordsPtr(), event.TitlePtr())
	return err
}
