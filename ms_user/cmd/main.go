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
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/client"
	"google.golang.org/grpc"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

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
	userService := service.NewUserService(userRepo)
	webhookService, err := service.NewWebhookService(userRepo, cfg.ClerkWebhookSecret)
	if err != nil {
		log.Fatalf("failed to create webhook service: %v", err)
	}
	authInterceptor := middleware.NewAuthInterceptor(clerkClient)
	grpcServer := server.NewGRPCServer(userService)

	// Start gRPC server
	go func() {
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
	}()

	// Start webhook server
	http.HandleFunc("/api/webhooks/clerk", webhookService.HandleWebhook)
	fmt.Printf("Webhook server listening on port %s\n", cfg.WebhookServerPort)
	if err := http.ListenAndServe(":"+cfg.WebhookServerPort, nil); err != nil {
		log.Fatalf("failed to serve webhook server: %v", err)
	}
}
