package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the system.
type User struct {
	ID                uuid.UUID
	ClerkUserID       string
	Email             string
	FullName          string
	Username          string
	Role              string // Role for authorization (admin, user, etc.)
	StorageUsedBytes  int64
	StorageQuotaBytes int64
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	Preferences       *UserPreferences
}

// UserPreferences holds settings for a user.
type UserPreferences struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	Theme                string
	Language             string
	Timezone             string
	CanvasSettings       map[string]interface{}
	NotificationSettings map[string]interface{}
	AccessibilitySettings map[string]interface{}
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// UserRepository defines the interface for interacting with user data.
type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByClerkID(ctx context.Context, clerkID string) (*User, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences *UserPreferences) (*UserPreferences, error)
}
