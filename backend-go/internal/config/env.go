package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"github.com/joho/godotenv"
)

func init() {
	// Determine project root based on file location
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Printf("Warning: unable to locate env.go path for .env loading")
		return
	}
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	envPath := filepath.Join(root, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: .env file not found at %s: %v", envPath, err)
	}
}

// GetEnv retrieves the value of the environment variable named by the key.
func GetEnv(key string) string {
	return os.Getenv(key)
}

// GetEnvOrDefault retrieves the value of the environment variable named by the key.
// If the variable is empty or not set, it returns the provided defaultValue.
func GetEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
