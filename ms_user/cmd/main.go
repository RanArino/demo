package main

import (
	"context"
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

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/client"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Initialize Clerk backend globally for the SDK
	clerk.SetBackend(clerk.NewBackend(&clerk.BackendConfig{
		HTTPClient: nil,
		Key:        &cfg.ClerkSecretKey,
	}))

	// Initialize Clerk client
	clerkClient := client.NewClient(&clerk.ClientConfig{
		BackendConfig: clerk.BackendConfig{
			Key: &cfg.ClerkSecretKey,
		},
	})

	// Create Ent client
	entClient, err := ent.Open("postgres", cfg.DSN)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer entClient.Close()

	// Run the auto migration tool.
	if err := entClient.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	// Initialize layers
	userRepo := repository.NewEntUserRepository(entClient)
	userService := service.NewUserService(userRepo, clerkClient)
	authInterceptor := middleware.NewAuthInterceptor(clerkClient, cfg.ClerkSecretKey)
	grpcServer := server.NewGRPCServer(userService)

	// Start HTTP server for webhooks in a separate goroutine
	go server.StartHTTPServer(cfg.WebhookServerPort, userService, cfg.ClerkWebhookSecret)

	// Start gRPC server
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
