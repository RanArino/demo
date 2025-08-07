package main

import (
	"demo/ms_knowledge/internal/config"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Load configuration for development
	cfg, err := config.LoadForDevelopment()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from the ms_knowledge service!")
	})

	port := ":" + cfg.GRPCPort
	fmt.Printf("ms_knowledge service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
