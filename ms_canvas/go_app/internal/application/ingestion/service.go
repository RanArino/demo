package ingestion

import (
	"time"

	"demo/ms_canvas/go_app/internal/domain"
	"demo/ms_canvas/go_app/internal/events"
	"demo/ms_canvas/go_app/internal/infrastructure/orchestrator/chunking"
	"github.com/google/uuid"
)

type Service struct {
	repo         domain.ContentRepository
	orchestrator chunking.Orchestrator
}

func NewService(repo domain.ContentRepository, orchestrator chunking.Orchestrator) *Service {
	return &Service{repo: repo, orchestrator: orchestrator}
}

func (s *Service) HandleDocumentProcessed(evt events.DocumentProcessedEvent) error {
	node := domain.ContentNode{
		ID:              uuid.New(),
		ContentSourceID: evt.ContentSourceID,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := s.repo.CreateContentNode(node); err != nil {
		return err
	}
	return s.orchestrator.TriggerChunking(evt.ContentSourceID)
}
