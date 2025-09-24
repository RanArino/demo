package service

import (
	"log"
)

// NOTE: This is a minimal placeholder task manager used by the
// EventOrchestrator and other components. The current service only
// consumes Kafka events (see internal/events/kafka/consumer.go). If the
// task execution flow should publish events to Kafka, implement a
// producer (e.g. internal/events/kafka/producer.go), expose a producer
// interface in internal/events, and inject that producer into the
// task executor so tasks can emit events (task.started/task.completed/etc.).
// Until those pieces are implemented this manager remains a local placeholder.
// SimpleTaskManager is a basic task manager for future use
type SimpleTaskManager struct {
	active bool
}

// NewTaskManager creates a simple task manager (placeholder)
func NewTaskManager() *SimpleTaskManager {
	log.Printf("[TaskManager] TaskManager infrastructure available for future event production")
	return &SimpleTaskManager{active: true}
}

// IsActive returns whether the task manager is active
func (stm *SimpleTaskManager) IsActive() bool {
	return stm.active
}
