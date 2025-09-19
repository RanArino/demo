package domain

import (
	"time"

	"github.com/google/uuid"
)

type ContentNode struct {
	ID              uuid.UUID
	ContentSourceID uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ContentRepository interface {
	CreateContentNode(node ContentNode) error
}
