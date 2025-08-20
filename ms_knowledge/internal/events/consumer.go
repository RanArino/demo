package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"demo/ms_knowledge/internal/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// ProcessedHandler handles DocumentProcessed events.
type ProcessedHandler interface {
	HandleDocumentProcessed(ctx context.Context, event DocumentProcessedEvent) error
}

// Consumer wraps a confluent-kafka-go consumer subscribed to document.processed.
type Consumer struct {
	client  *kafka.Consumer
	handler ProcessedHandler
}

// NewConsumer creates a consumer configured for Confluent Cloud and subscribed to the processed topic.
func NewConsumer(cfg *config.Config, groupID string, handler ProcessedHandler) (*Consumer, error) {
	conf := kafka.ConfigMap{
		"bootstrap.servers":        cfg.Kafka.Brokers,
		"security.protocol":        cfg.Kafka.SecurityProtocol,
		"sasl.mechanisms":          "PLAIN",
		"sasl.username":            cfg.Kafka.SaslUsername,
		"sasl.password":            cfg.Kafka.SaslPassword,
		"group.id":                 groupID,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false, // Manual commit for at-least-once delivery
		"session.timeout.ms":       45000,
		"heartbeat.interval.ms":    3000,
		"max.poll.interval.ms":     600000, // 10 minutes
		"isolation.level":          "read_committed",
	}
	client, err := kafka.NewConsumer(&conf)
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}
	if err := client.SubscribeTopics([]string{TopicDocumentProcessed}, nil); err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	return &Consumer{client: client, handler: handler}, nil
}

// Close closes the consumer.
func (c *Consumer) Close() {
	if c == nil || c.client == nil {
		return
	}
	_ = c.client.Close()
}

// Run starts the polling loop until ctx is done.
func (c *Consumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		e := c.client.Poll(1000)
		switch ev := e.(type) {
		case *kafka.Message:
			var payload DocumentProcessedEvent
			if err := json.Unmarshal(ev.Value, &payload); err != nil {
				log.Printf("failed to unmarshal event: %v", err)
				continue
			}
			if err := c.handler.HandleDocumentProcessed(ctx, payload); err != nil {
				log.Printf("handler error: %v", err)
			}
		case kafka.Error:
			log.Printf("kafka error: %v", ev)
		}
	}
}
