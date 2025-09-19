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
	Neo4jVectorDimensions int
	Topics                Topics
	PythonService         PythonServiceConfig
}

type PythonServiceConfig struct {
	Host        string
	Port        int
	MaxMsgBytes int
	BatchSize   int
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
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
		GRPCPort:              getEnvInt("GRPC_PORT", 50055), // Different from Python port
		ClerkSecretKey:        getEnv("CLERK_SECRET_KEY", ""),
		Neo4jURI:              getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUsername:         getEnv("NEO4J_USERNAME", "neo4j"),
		Neo4jPassword:         getEnv("NEO4J_PASSWORD", "password"),
		Neo4jDatabase:         getEnv("NEO4J_DATABASE", "neo4j"),
		Neo4jVectorDimensions: getEnvInt("NEO4J_VECTOR_DIMENSIONS", 1536),
		Topics: Topics{
			DocumentProcessed: getEnv("TOPIC_DOCUMENT_PROCESSED", "document.processed"),
		},
		PythonService: PythonServiceConfig{
			Host:        getEnv("CANVAS_PY_HOST", "0.0.0.0"),
			Port:        getEnvInt("CANVAS_PY_PORT", 50054),
			MaxMsgBytes: getEnvInt("CANVAS_MAX_GRPC_MSG_BYTES", 64*1024*1024),
			BatchSize:   getEnvInt("CANVAS_BATCH_SIZE", 50),
		},
	}
}
