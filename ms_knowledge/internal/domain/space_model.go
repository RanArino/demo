package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Space represents a knowledge space
type Space struct {
	ID                 uuid.UUID  `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Icon               string     `json:"icon"`
	CoverImage         string     `json:"cover_image"`
	Keywords           []string   `json:"keywords"`
	OwnerID            uuid.UUID  `json:"owner_id"`
	CreatedAt          time.Time  `json:"created_at"`
	CreatedBy          uuid.UUID  `json:"created_by"`
	LastUpdatedAt      time.Time  `json:"last_updated_at"`
	LastUpdatedBy      uuid.UUID  `json:"last_updated_by"`
	AccessLevel        string     `json:"access_level"`
	GuestAccessEnabled bool       `json:"guest_access_enabled"`
	GuestAccessExpiry  *time.Time `json:"guest_access_expiry"`
	Status             string     `json:"status"`
	ProcessingStatus   string     `json:"processing_status"`
	DeletedAt          *time.Time `json:"deleted_at"`
}

// SpaceStats represents aggregated statistics for a space
type SpaceStats struct {
	ContentCount    int64                  `json:"content_count"`
	LinkCount       int64                  `json:"link_count"`
	TotalSizeBytes  int64                  `json:"total_size_bytes"`
	LastActivityAt  time.Time              `json:"last_activity_at"`
	ContentByStatus map[string]int64       `json:"content_by_status"`
	ProcessingStats ContentProcessingStats `json:"processing_stats"`
}

// ContentProcessingStats represents processing-related statistics
type ContentProcessingStats struct {
	UploadingCount  int64 `json:"uploading_count"`
	UploadedCount   int64 `json:"uploaded_count"`
	ProcessingCount int64 `json:"processing_count"`
	ProcessedCount  int64 `json:"processed_count"`
	FailedCount     int64 `json:"failed_count"`
	PendingCount    int64 `json:"pending_count"`
}

// SpaceWithStats represents a space with its statistics
type SpaceWithStats struct {
	Space
	Stats SpaceStats `json:"stats"`
}

// SpaceFilter represents filters for listing spaces
type SpaceFilter struct {
	OwnerID       uuid.UUID
	Keywords      []string
	Query         string
	AccessLevel   string
	Status        string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
	Limit         int
	Offset        int
}

// SpaceRepository defines the interface for space data operations
type SpaceRepository interface {
	Create(ctx context.Context, space *Space) error
	GetByID(ctx context.Context, id uuid.UUID) (*Space, error)
	GetWithStats(ctx context.Context, id uuid.UUID) (*SpaceWithStats, error)
	List(ctx context.Context, filter SpaceFilter) ([]*Space, error)
	ListWithStats(ctx context.Context, filter SpaceFilter) ([]*SpaceWithStats, error)
	Update(ctx context.Context, space *Space) error
	UpdateStats(ctx context.Context, spaceID uuid.UUID, docCountChange int, sizeChange int64) error
	Delete(ctx context.Context, id uuid.UUID, hardDelete bool) error
	Search(ctx context.Context, query string, filter SpaceFilter) ([]*Space, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

// Context keys used across services
type ContextKey string

const (
	// OwnerIDKey carries the resolved internal user UUID as string in context
	OwnerIDKey ContextKey = "owner_id"
	// RoleKey carries the caller's role (e.g., admin, user) from JWT
	RoleKey ContextKey = "role"
	// ClerkUserIDKey carries the Clerk user ID from JWT (for debugging/logging)
	ClerkUserIDKey ContextKey = "clerk_user_id"
)

// GetOwnerID retrieves the internal user ID from context
func GetOwnerID(ctx context.Context) (string, bool) {
	ownerID, ok := ctx.Value(OwnerIDKey).(string)
	return ownerID, ok && ownerID != ""
}

// GetRole retrieves the user role from context
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok && role != ""
}

// GetClerkUserID retrieves the Clerk user ID from context
func GetClerkUserID(ctx context.Context) (string, bool) {
	clerkUserID, ok := ctx.Value(ClerkUserIDKey).(string)
	return clerkUserID, ok && clerkUserID != ""
}

// IsAdmin checks if the user's role in the context is 'admin'
func IsAdmin(ctx context.Context) bool {
	role, ok := GetRole(ctx)
	return ok && strings.ToLower(role) == "admin"
}
