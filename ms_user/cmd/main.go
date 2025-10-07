package main

import (
	"context"
	"database/sql"
	userv1 "demo/ms_user/api/proto/v1"
	"demo/ms_user/ent"
	"demo/ms_user/internal/config"
	"demo/ms_user/internal/middleware"
	"demo/ms_user/internal/repository"
	"demo/ms_user/internal/server"
	"demo/ms_user/internal/service"
	"fmt"
	"log"
	"net"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/client"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	clerk.SetBackend(clerk.NewBackend(&clerk.BackendConfig{
		HTTPClient: nil,
		Key:        &cfg.ClerkSecretKey,
	}))

	clerkClient := client.NewClient(&clerk.ClientConfig{
		BackendConfig: clerk.BackendConfig{
			Key: &cfg.ClerkSecretKey,
		},
	})

	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.Postgres, db)
	entClient := ent.NewClient(ent.Driver(drv))
	defer entClient.Close()

	if err := entClient.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	userRepo := repository.NewEntUserRepository(entClient)
	userService := service.NewUserService(userRepo, clerkClient, cfg.DefaultStorageQuotaGB)
	authInterceptor := middleware.NewAuthInterceptor(clerkClient, cfg.ClerkSecretKey)
	grpcServer := server.NewGRPCServer(userService)

	go server.StartHTTPServer(cfg.WebhookServerPort, userService, cfg.ClerkWebhookSecret)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor.Unary()))
	userv1.RegisterUserServiceServer(s, grpcServer)

	fmt.Printf("gRPC server listening on port %s\n", cfg.GRPCServerPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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

	log.Printf("Successfully connected to database at %s:%d", pgxConfig.Host, pgxConfig.Port)
	return db, nil
}
