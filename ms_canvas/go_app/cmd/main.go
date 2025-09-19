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
	python "demo/ms_canvas/go_app/internal/gateway/python"
	kafka "demo/ms_canvas/go_app/internal/infrastructure/consumer/kafka"
	neo "demo/ms_canvas/go_app/internal/infrastructure/repository/neo4j"
	"demo/ms_canvas/go_app/internal/workflows"
)

func main() {
	cfg := config.Load()

	drv, err := neo.NewDriver(cfg.Neo4jURI, cfg.Neo4jUsername, cfg.Neo4jPassword, cfg.Neo4jDatabase)
	if err != nil {
		log.Fatalf("failed to create neo4j driver: %v", err)
	}
	defer drv.Close(context.Background())
	_ = drv.EnsureConstraints(context.Background())
	_ = drv.EnsureIndexes(context.Background())

	nodeRepo := neo.NewNodeRepo(drv)
	linkRepo := neo.NewLinkRepo(drv)

	chunkingClient := python.NewChunkingClient(cfg, nodeRepo, linkRepo)
	documentWorkflow := workflows.NewDocumentWorkflow(nodeRepo, chunkingClient)
	eventHandler := workflows.NewEventHandler(documentWorkflow)

	consumer, err := kafka.NewConsumer(cfg, eventHandler)
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("consumer stopped with error: %v", err)
		}
	}()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})
	port := ":8080"
	fmt.Printf("ms_canvas service starting on port %s\n", port)
	server := &http.Server{Addr: port}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Fatal(server.ListenAndServe())
}
