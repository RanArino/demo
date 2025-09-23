package kafka

import (
	"context"
	"encoding/json"
	"log"

	"demo/ms_canvas/go_app/internal/config"
	"demo/ms_canvas/go_app/internal/events"
	"demo/ms_canvas/go_app/internal/service"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer struct {
	inner   *ckafka.Consumer
	topic   string
	handler *Handler
}

func NewConsumer(cfg config.Config, eventHandler service.EventHandler) (*Consumer, error) {
	conf := &ckafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
		"group.id":          cfg.KafkaGroupID,
		"auto.offset.reset": "earliest",
		"security.protocol": cfg.KafkaSecurityProtocol,
	}
	if cfg.KafkaSecurityProtocol == "SASL_SSL" {
		_ = conf.SetKey("sasl.mechanisms", "PLAIN")
		_ = conf.SetKey("sasl.username", cfg.KafkaSaslUsername)
		_ = conf.SetKey("sasl.password", cfg.KafkaSaslPassword)
		_ = conf.SetKey("ssl.endpoint.identification.algorithm", "https")
	}
	c, err := ckafka.NewConsumer(conf)
	if err != nil {
		return nil, err
	}
	h := NewHandler(cfg, eventHandler)
	return &Consumer{inner: c, topic: cfg.Topics.DocumentProcessed, handler: h}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	if err := c.inner.Subscribe(c.topic, nil); err != nil {
		return err
	}
	log.Printf("[Kafka] Subscribed to %s", c.topic)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[Kafka] Shutting down consumer")
			_ = c.inner.Close()
			return nil
		default:
			ev := c.inner.Poll(100)
			switch e := ev.(type) {
			case *ckafka.Message:
				if e.TopicPartition.Error != nil {
					log.Printf("[Kafka] partition error: %v", e.TopicPartition.Error)
					continue
				}
				var msg events.DocumentProcessedEvent
				if err := json.Unmarshal(e.Value, &msg); err != nil {
					log.Printf("[Kafka] json decode error: %v", err)
					continue
				}
				headers := map[string]string{}
				for _, hdr := range e.Headers {
					headers[hdr.Key] = string(hdr.Value)
				}
				if err := c.handler.Handle(msg, headers); err != nil {
					log.Printf("[Handler] error: %v", err)
					continue
				}
			case ckafka.Error:
				log.Printf("[Kafka] error: %v", e)
			default:
				// ignore other events
			}
		}
	}
}
