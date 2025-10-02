package config

import (
	"os"
	"strconv"
)

type Topics struct {
	DocumentProcessed string
}

type Config struct {
	KafkaBrokers                  string
	KafkaSecurityProtocol         string
	KafkaSaslUsername             string
	KafkaSaslPassword             string
	KafkaGroupID                  string
	GRPCPort                      int
	ClerkSecretKey                string
	GEMINIAPIKey                  string
	Neo4jURI                      string
	Neo4jUsername                 string
	Neo4jPassword                 string
	Neo4jDatabase                 string
	Neo4jVectorDimensions         int
	Neo4jConstraintTimeoutSeconds int
	Neo4jIndexTimeoutSeconds      int
	Neo4jEnsureMaxRetries         int
	Neo4jEnsureRetryDelaySeconds  int
	Topics                        Topics
	PythonService                 PythonServiceConfig
	R2Config                      R2Config

	// OpenAI embedding configuration for Neo4j GenAI
	OpenAIAPIKey         string
	OpenAIEmbeddingModel string
	OpenAIEmbeddingDim   int
}

type R2Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	AccountID       string
	BucketProcessed string
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
		KafkaBrokers:                  getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaSecurityProtocol:         getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
		KafkaSaslUsername:             getEnv("KAFKA_SASL_USERNAME", ""),
		KafkaSaslPassword:             getEnv("KAFKA_SASL_PASSWORD", ""),
		KafkaGroupID:                  getEnv("KAFKA_GROUP_ID", "ms_canvas-consumer-group"),
		GRPCPort:                      getEnvInt("GRPC_PORT", 50055), // Different from Python port
		ClerkSecretKey:                getEnv("CLERK_SECRET_KEY", ""),
		GEMINIAPIKey:                  getEnv("GEMINI_API_KEY", ""),
		Neo4jURI:                      getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUsername:                 getEnv("NEO4J_USERNAME", "neo4j"),
		Neo4jPassword:                 getEnv("NEO4J_PASSWORD", "password"),
		Neo4jDatabase:                 getEnv("NEO4J_DATABASE", "neo4j"),
		Neo4jVectorDimensions:         getEnvInt("NEO4J_VECTOR_DIMENSIONS", 1536),
		Neo4jConstraintTimeoutSeconds: getEnvInt("NEO4J_CONSTRAINT_TIMEOUT_SECONDS", 15),
		Neo4jIndexTimeoutSeconds:      getEnvInt("NEO4J_INDEX_TIMEOUT_SECONDS", 30),
		Neo4jEnsureMaxRetries:         getEnvInt("NEO4J_ENSURE_MAX_RETRIES", 2),
		Neo4jEnsureRetryDelaySeconds:  getEnvInt("NEO4J_ENSURE_RETRY_DELAY_SECONDS", 5),
		Topics: Topics{
			DocumentProcessed: getEnv("TOPIC_DOCUMENT_PROCESSED", "document.processed"),
		},
		PythonService: PythonServiceConfig{
			Host:        getEnv("CANVAS_PY_HOST", "0.0.0.0"),
			Port:        getEnvInt("CANVAS_PY_PORT", 50054),
			MaxMsgBytes: getEnvInt("CANVAS_MAX_GRPC_MSG_BYTES", 64*1024*1024),
			BatchSize:   getEnvInt("CANVAS_BATCH_SIZE", 50),
		},
		R2Config: R2Config{
			Endpoint:        getEnv("R2_ENDPOINT", ""),
			AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
			AccountID:       getEnv("R2_ACCOUNT_ID", ""),
			BucketProcessed: getEnv("R2_BUCKET_PROCESSED", ""),
		},

		// OpenAI embedding configuration
		OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
		OpenAIEmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		OpenAIEmbeddingDim:   getEnvInt("OPENAI_EMBEDDING_DIM", 0),
	}
}
