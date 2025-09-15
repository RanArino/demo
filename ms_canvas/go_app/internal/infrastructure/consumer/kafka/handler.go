package kafka

import (
	"errors"
	"log"
	"strings"

	ingestion "demo/ms_canvas/go_app/internal/application/ingestion"
	"demo/ms_canvas/go_app/internal/config"
	"demo/ms_canvas/go_app/internal/events"
)

type Handler struct {
	cfg config.Config
	svc *ingestion.Service
}

func NewHandler(cfg config.Config, svc *ingestion.Service) *Handler {
	return &Handler{cfg: cfg, svc: svc}
}

func (h *Handler) ValidateAndNormalize(evt events.DocumentProcessedEvent, headers map[string]string) (events.DocumentProcessedEvent, error) {
	if strings.TrimSpace(evt.Status) == "" {
		return evt, errors.New("missing status")
	}
	if strings.ToUpper(evt.Status) != "PROCESSED" {
		log.Printf("[Handler] skipping event with status=%s", evt.Status)
		return evt, nil
	}
	return evt, nil
}

func (h *Handler) Handle(evt events.DocumentProcessedEvent, headers map[string]string) error {
	norm, err := h.ValidateAndNormalize(evt, headers)
	if err != nil {
		return err
	}
	if strings.ToUpper(norm.Status) != "PROCESSED" {
		return nil
	}
	return h.svc.HandleDocumentProcessed(norm)
}
