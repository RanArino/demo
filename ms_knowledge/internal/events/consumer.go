package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
	log.Printf("Starting consumer loop...")
	defer log.Printf("Consumer loop stopped")

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping consumer")
			return
		default:
		}

		e := c.client.Poll(1000)
		if e == nil {
			continue
		}

		switch ev := e.(type) {
		case *kafka.Message:
			c.handleMessage(ctx, ev)
		case kafka.Error:
			if ev.Code() == kafka.ErrAllBrokersDown {
				log.Printf("All brokers down, retrying in 5 seconds: %v", ev)
				time.Sleep(5 * time.Second)
				continue
			}
			log.Printf("Kafka error: %v", ev)
		case *kafka.OffsetsCommitted:
			if ev.Error != nil {
				log.Printf("Failed to commit offsets: %v", ev.Error)
			}
		default:
			log.Printf("Ignored event: %v", e)
		}
	}
}

// handleMessage processes a single Kafka message with retry logic and idempotency
func (c *Consumer) handleMessage(ctx context.Context, msg *kafka.Message) {
	var payload DocumentProcessedEvent
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		log.Printf("Failed to unmarshal event (invalid JSON, discarding): %v", err)
		c.commitMessage(msg)
		return
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	baseDelay := time.Second
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := c.handler.HandleDocumentProcessed(ctx, payload); err != nil {
			if attempt == maxRetries {
				log.Printf("Failed to process event after %d attempts, content_source_id=%s, error: %v", 
					maxRetries+1, payload.ContentSourceID, err)
				// TODO: Send to dead letter queue or alert system
				break
			}
			
			delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
			log.Printf("Handler error (attempt %d/%d), retrying in %v: %v", 
				attempt+1, maxRetries+1, delay, err)
			
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
				continue
			}
		} else {
			// Success - break retry loop
			log.Printf("Successfully processed event for content_source_id=%s", payload.ContentSourceID)
			break
		}
	}
	
	// Commit the message offset after processing (or failing completely)
	c.commitMessage(msg)
}

// commitMessage commits the message offset
func (c *Consumer) commitMessage(msg *kafka.Message) {
	_, err := c.client.CommitMessage(msg)
	if err != nil {
		log.Printf("Failed to commit message: %v", err)
	}
}
