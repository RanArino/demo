package middleware

import (
	"context"
	"demo/ms_user/internal/domain"
	"strings"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// UserIDKey is the context key for the user's Clerk ID.
	UserIDKey ContextKey = "user_id"
	// UserRoleKey is the context key for the user's role.
	UserRoleKey ContextKey = "user_role"
	// UsernameKey is the context key for the user's username.
	UsernameKey ContextKey = "username"
	// UserKey is the context key for the user's domain model object.
	UserKey ContextKey = "user"
)

// IsAdmin checks if the user's role in the context is 'admin'.
func IsAdmin(ctx context.Context) bool {
	role, ok := ctx.Value(UserRoleKey).(string)
	return ok && strings.ToLower(role) == "admin"
}

// GetCurrentUser retrieves the user object from the context.
func GetCurrentUser(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(UserKey).(*domain.User)
	return user, ok
}
