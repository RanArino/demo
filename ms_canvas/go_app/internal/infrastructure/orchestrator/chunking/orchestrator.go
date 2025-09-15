package chunking

import (
	"log"
	"github.com/google/uuid"
)

type Orchestrator interface {
	TriggerChunking(contentSourceID uuid.UUID) error
}

type NoopOrchestrator struct{}

func NewNoopOrchestrator() *NoopOrchestrator { return &NoopOrchestrator{} }

func (o *NoopOrchestrator) TriggerChunking(contentSourceID uuid.UUID) error {
	log.Printf("[Chunking] TriggerChunking for content_source_id=%s", contentSourceID.String())
	return nil
}
