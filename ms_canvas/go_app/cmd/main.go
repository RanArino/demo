package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"demo/ms_canvas/go_app/internal/config"
	"demo/ms_canvas/go_app/internal/events/kafka"
	"demo/ms_canvas/go_app/internal/gateway/python"
	"demo/ms_canvas/go_app/internal/repository/neo4j"
	"demo/ms_canvas/go_app/internal/service"
)

func main() {
	cfg := config.Load()
	log.Printf("[Main] Starting ms_canvas service with Kafka consumer for event-streaming")

	// Initialize Neo4j driver and repositories
	drv, err := neo4j.NewDriver(cfg.Neo4jURI, cfg.Neo4jUsername, cfg.Neo4jPassword, cfg.Neo4jDatabase, &neo4j.DriverOptions{VectorDimensions: cfg.Neo4jVectorDimensions})
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

	port := fmt.Sprintf(":%d", cfg.GRPCPort)
	log.Printf("[Main] ms_canvas service starting on port %s", port)
	log.Printf("[Main] Ready to receive Kafka events from topic: %s", cfg.Topics.DocumentProcessed)

	server := &http.Server{Addr: port}

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		log.Printf("[Main] Shutting down...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("[Main] Server shutdown error: %v", err)
		}

		log.Printf("[Main] Shutdown complete")
	}()

	log.Fatal(server.ListenAndServe())
}
