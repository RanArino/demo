package service

import (
	"context"
	"fmt"
	"strings"

	"demo/ms_knowledge/internal/domain"

	"github.com/google/uuid"
)

// MustGetOwnerUUID extracts and validates the caller's owner UUID from context.
func MustGetOwnerUUID(ctx context.Context) (uuid.UUID, error) {
	ownerIDStr, ok := ctx.Value(domain.OwnerIDKey).(string)
	if !ok || strings.TrimSpace(ownerIDStr) == "" {
		return uuid.Nil, fmt.Errorf("owner_id missing in context")
	}
	ownerUUID, err := uuid.Parse(strings.TrimSpace(ownerIDStr))
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid owner_id format: %w", err)
	}
	return ownerUUID, nil
}

// EnforceOwner ensures the caller (from context) matches the given resource owner.
func EnforceOwner(ctx context.Context, resourceOwner uuid.UUID) error {
	caller, err := MustGetOwnerUUID(ctx)
	if err != nil {
		return err
	}
	if resourceOwner != caller {
		return fmt.Errorf("forbidden: not the owner")
	}
	return nil
}
