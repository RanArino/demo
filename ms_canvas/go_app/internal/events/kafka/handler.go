package kafka

import (
	"errors"
	"log"
	"strings"

	"demo/ms_canvas/go_app/internal/config"
	"demo/ms_canvas/go_app/internal/events"
	"demo/ms_canvas/go_app/internal/service"
)

type Handler struct {
	cfg          config.Config
	eventHandler service.EventHandler
}

func NewHandler(cfg config.Config, eventHandler service.EventHandler) *Handler {
	return &Handler{cfg: cfg, eventHandler: eventHandler}
}

func (h *Handler) ValidateAndNormalize(evt events.DocumentProcessedEvent, headers map[string]string) (events.DocumentProcessedEvent, error) {
	status := strings.TrimSpace(string(evt.Status))
	if status == "" {
		return evt, errors.New("missing status")
	}

	normalized := events.ProcessStatus(strings.ToUpper(status))
	if normalized != events.ProcessStatusProcessed {
		log.Printf("[Handler] skipping event with status=%s", normalized)
		return evt, nil
	}

	evt.Status = normalized
	return evt, nil
}

func (h *Handler) Handle(evt events.DocumentProcessedEvent, headers map[string]string) error {
	norm, err := h.ValidateAndNormalize(evt, headers)
	if err != nil {
		return err
	}
	if norm.Status != events.ProcessStatusProcessed {
		return nil
	}
	return h.eventHandler.HandleDocumentProcessed(norm)
}
