package events

import (
	"context"
	"encoding/json"
	"fmt"

	"demo/ms_knowledge/internal/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Producer is a thin wrapper around confluent-kafka-go producer configured from our config.
type Producer struct {
	client *kafka.Producer
}

// NewProducer builds a new Kafka producer using Confluent Cloud SASL/SSL configuration
func NewProducer(cfg *config.Config) (*Producer, error) {
	conf := kafka.ConfigMap{
		"bootstrap.servers":        cfg.Kafka.Brokers,
		"security.protocol":        cfg.Kafka.SecurityProtocol, // SASL_SSL
		"sasl.mechanisms":          "PLAIN",
		"sasl.username":            cfg.Kafka.SaslUsername,
		"sasl.password":            cfg.Kafka.SaslPassword,
		"acks":                     "all",      // Wait for all in-sync replicas
		"retries":                  "10",       // Retry failed sends
		"retry.backoff.ms":         "100",      // Backoff between retries
		"request.timeout.ms":       "30000",    // 30 second timeout
		"delivery.timeout.ms":      "300000",   // 5 minute total delivery timeout
		"max.in.flight.requests.per.connection": "5", // Pipeline for performance
		"enable.idempotence":       "true",     // Exactly-once semantics
	}

	client, err := kafka.NewProducer(&conf)
	if err != nil {
		return nil, fmt.Errorf("create producer: %w", err)
	}
	return &Producer{client: client}, nil
}

// Close flushes pending messages and closes the producer.
func (p *Producer) Close() {
	if p == nil || p.client == nil {
		return
	}
	_ = p.client.Flush(15000)
	p.client.Close()
}

// ProduceJSON marshals v as JSON and produces it to topic.
func (p *Producer) ProduceJSON(ctx context.Context, topic string, key string, v any) error {
	if p == nil || p.client == nil {
		return fmt.Errorf("producer not initialized")
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          bytes,
	}
	return p.client.Produce(msg, nil)
}
