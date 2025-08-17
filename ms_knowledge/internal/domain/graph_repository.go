package domain

import (
	"context"

	"github.com/google/uuid"
)

type GraphRepository interface {
	CreateContentNode(ctx context.Context, contentID uuid.UUID) error
	CreateLink(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID) error
	CreateLinkWithType(ctx context.Context, fromContentID uuid.UUID, toContentID uuid.UUID, relationshipType string) error
	GetBacklinks(ctx context.Context, contentID uuid.UUID) ([]uuid.UUID, error)
}
