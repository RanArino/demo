package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server      ServerConfig
	ML          MLServiceConfig
	Qdrant      QdrantConfig
	GeminiAPI   GeminiConfig
	VectorStore VectorStoreConfig
	LLM         LLMConfig
}

// ServerConfig holds configuration for the HTTP server
type ServerConfig struct {
	Port int
}

// MLServiceConfig holds configuration for the Python ML service
type MLServiceConfig struct {
	BaseURL string
}

// QdrantConfig holds configuration for Qdrant vector database
type QdrantConfig struct {
	Host   string
	Port   int
	APIKey string
	UseTLS bool
}

// GeminiConfig holds configuration for Google's Gemini API
type GeminiConfig struct {
	APIKey string
}

// VectorStoreConfig holds Qdrant-specific configuration
type VectorStoreConfig struct {
	Endpoint    string
	APIKey      string
	Collections struct {
		Summaries string
		Chunks    string
	}
}

// LLMConfig holds Gemini-specific configuration
type LLMConfig struct {
	APIKey           string
	EmbeddingModel   string
	EmbeddingDim     uint64
	MaxTokensPerCall int
}

// Default collection names
const (
	DefaultSummariesCollection = "doc_summaries"
	DefaultChunksCollection    = "doc_chunks"
)

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Set defaults
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		ML: MLServiceConfig{
			BaseURL: "http://backend-py:5000",
		},
		Qdrant: QdrantConfig{
			Host:   "qdrant",
			Port:   6333,
			APIKey: "",
			UseTLS: true,
		},
	}

	// Override with environment variables if present
	if port := os.Getenv("SERVER_PORT"); port != "" {
		portInt, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid SERVER_PORT value: %v", err)
		}
		cfg.Server.Port = portInt
	}

	if mlURL := os.Getenv("ML_SERVICE_URL"); mlURL != "" {
		cfg.ML.BaseURL = mlURL
	}

	if qdrantHost := os.Getenv("QDRANT_HOST"); qdrantHost != "" {
		cfg.Qdrant.Host = qdrantHost
	}

	if qdrantPort := os.Getenv("QDRANT_PORT"); qdrantPort != "" {
		portInt, err := strconv.Atoi(qdrantPort)
		if err != nil {
			return nil, fmt.Errorf("invalid QDRANT_PORT value: %v", err)
		}
		cfg.Qdrant.Port = portInt
	}
	if qdrantApiKey := os.Getenv("QDRANT_API_KEY"); qdrantApiKey != "" {
		cfg.Qdrant.APIKey = qdrantApiKey
	}

	if geminiAPIKey := os.Getenv("GEMINI_API_KEY"); geminiAPIKey != "" {
		cfg.GeminiAPI.APIKey = geminiAPIKey
	} else {
		// For demo purposes, we'll allow running without a Gemini API key
		// but in production, this should be required
		fmt.Println("Warning: GEMINI_API_KEY not set")
	}

	// Vector store config
	cfg.VectorStore.Endpoint = getEnvOrDefault("QDRANT_ENDPOINT", "http://localhost:6334")
	cfg.VectorStore.APIKey = os.Getenv("QDRANT_API_KEY")
	cfg.VectorStore.Collections.Summaries = DefaultSummariesCollection
	cfg.VectorStore.Collections.Chunks = DefaultChunksCollection

	// LLM config
	cfg.LLM.APIKey = os.Getenv("GEMINI_API_KEY")
	cfg.LLM.EmbeddingModel = getEnvOrDefault("GEMINI_EMBEDDING_MODEL", "models/embedding-001")
	cfg.LLM.EmbeddingDim = 768 // Default dimension for Gemini embeddings
	cfg.LLM.MaxTokensPerCall = 1024

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
