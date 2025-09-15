package config

import (
	"os"
	"strconv"
)

type Topics struct {
	DocumentProcessed string
}

type Config struct {
	KafkaBrokers          string
	KafkaSecurityProtocol string
	KafkaSaslUsername     string
	KafkaSaslPassword     string
	KafkaGroupID          string
	GRPCPort              int
	ClerkSecretKey        string
	Neo4jURI              string
	Neo4jUsername         string
	Neo4jPassword         string
	Neo4jDatabase         string
	Topics                Topics
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	return Config{
		KafkaBrokers:          getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaSecurityProtocol: getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
		KafkaSaslUsername:     getEnv("KAFKA_SASL_USERNAME", ""),
		KafkaSaslPassword:     getEnv("KAFKA_SASL_PASSWORD", ""),
		KafkaGroupID:          getEnv("KAFKA_GROUP_ID", "ms_canvas-consumer-group"),
		GRPCPort:              func() int { if v := os.Getenv("GRPC_PORT"); v != "" { if n, err := strconv.Atoi(v); err == nil { return n } } ; return 50051 }(),
		ClerkSecretKey:        getEnv("CLERK_SECRET_KEY", ""),
		Neo4jURI:              getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUsername:         getEnv("NEO4J_USERNAME", "neo4j"),
		Neo4jPassword:         getEnv("NEO4J_PASSWORD", "password"),
		Neo4jDatabase:         getEnv("NEO4J_DATABASE", "neo4j"),
		Topics: Topics{
			DocumentProcessed: getEnv("TOPIC_DOCUMENT_PROCESSED", "document.processed"),
		},
	}
}
