package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	canvasv1 "demo/ms_canvas/go_app/api/proto/public/v1"
	"demo/ms_canvas/go_app/internal/config"
	"demo/ms_canvas/go_app/internal/events/kafka"
	"demo/ms_canvas/go_app/internal/gateway/python"
	"demo/ms_canvas/go_app/internal/repository/neo4j"
	"demo/ms_canvas/go_app/internal/server"
	"demo/ms_canvas/go_app/internal/service"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	log.Printf("[Main] Starting ms_canvas service with Kafka consumer for event-streaming")

	// Initialize Neo4j driver and repositories
	drv, err := neo4j.NewDriver(
		cfg.Neo4jURI,
		cfg.Neo4jUsername,
		cfg.Neo4jPassword,
		cfg.Neo4jDatabase,
		&neo4j.DriverOptions{
			VectorDimensions:  cfg.Neo4jVectorDimensions,
			ConstraintTimeout: time.Duration(cfg.Neo4jConstraintTimeoutSeconds) * time.Second,
			IndexTimeout:      time.Duration(cfg.Neo4jIndexTimeoutSeconds) * time.Second,
			MaxEnsureRetries:  cfg.Neo4jEnsureMaxRetries,
			EnsureRetryDelay:  time.Duration(cfg.Neo4jEnsureRetryDelaySeconds) * time.Second,
		},
	)
	if err != nil {
		log.Fatalf("failed to create neo4j driver: %v", err)
	}
	defer drv.Close(context.Background())

	// Ensure Neo4j constraints and indexes are created
	if err := drv.EnsureConstraints(context.Background()); err != nil {
		log.Printf("[Main] Warning: failed to ensure Neo4j constraints: %v", err)
	}

	nodeRepo := neo4j.NewNodeRepo(drv)
	linkRepo := neo4j.NewLinkRepo(drv)
	searchRepo := neo4j.NewSearchRepo(drv)

	// Initialize Python gateway for ML operations
	gatewayFactory := python.NewGatewayFactory(cfg)
	pythonGateway, err := gatewayFactory.GetChunkingEmbeddingGateway()
	if err != nil {
		log.Fatalf("failed to create Python gateway: %v", err)
	}
	defer gatewayFactory.Close()

	// Create EventOrchestrator to handle document ingestion events
	eventOrchestrator := service.NewEventOrchestrator(
		cfg,
		nodeRepo,
		linkRepo,
		pythonGateway,
	)

	// Initialize services for gRPC server
	searchService := service.NewSearchService(pythonGateway, searchRepo)
	nodeService := service.NewNodeService(nodeRepo, linkRepo)
	linkService := service.NewLinkService(linkRepo)

	// Create Kafka consumer with the EventOrchestrator as event handler
	consumer, err := kafka.NewConsumer(cfg, eventOrchestrator)
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}

	// Set up graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start Kafka consumer in background
	log.Printf("[Main] Starting Kafka consumer for topic: %s", cfg.Topics.DocumentProcessed)
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("[Main] Consumer stopped with error: %v", err)
		}
	}()

	// Initialize gRPC server for CanvasPublic API
	grpcSrv := grpc.NewServer()
	canvasServer := server.NewCanvasPublicServer(searchService, nodeService, linkService)
	canvasv1.RegisterCanvasPublicServer(grpcSrv, canvasServer)

	// Start gRPC server on configured port (50054 per requirements)
	grpcListener, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("failed to listen on gRPC port 50054: %v", err)
	}
	go func() {
		log.Printf("[Main] gRPC server listening on :50054")
		if err := grpcSrv.Serve(grpcListener); err != nil {
			log.Printf("[Main] gRPC server stopped: %v", err)
		}
	}()

	// Set up health check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	// Set up readiness endpoint
	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Check if all components are ready
		healthResults := gatewayFactory.HealthCheck(ctx)
		ready := len(healthResults) > 0

		if ready {
			fmt.Fprint(w, "ready")
		} else {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
		}
	})

	httpPort := ":8080"
	log.Printf("[Main] ms_canvas service starting - gRPC on :50054, HTTP health checks on %s", httpPort)
	log.Printf("[Main] Ready to receive Kafka events from topic: %s", cfg.Topics.DocumentProcessed)

	server := &http.Server{Addr: httpPort}

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		log.Printf("[Main] Shutting down...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Shutdown gRPC server gracefully
		log.Printf("[Main] Stopping gRPC server...")
		grpcSrv.GracefulStop()

		// Shutdown HTTP server
		log.Printf("[Main] Stopping HTTP server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("[Main] HTTP server shutdown error: %v", err)
		}

		log.Printf("[Main] Shutdown complete")
	}()

	log.Fatal(server.ListenAndServe())
}
