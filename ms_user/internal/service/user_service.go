package service

import (
	"context"
	"demo/ms_user/internal/domain"
	"demo/ms_user/internal/middleware"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/clerk/clerk-sdk-go/v2/client"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/google/uuid"
)

// UserService provides user-related business logic.
type UserService struct {
	repo                   domain.UserRepository
	clerkClient            *client.Client
	defaultStorageQuotaBytes int64
}

// NewUserService creates a new UserService.
func NewUserService(repo domain.UserRepository, clerkClient *client.Client, defaultStorageQuotaGB int64) *UserService {
	return &UserService{
		repo:                   repo,
		clerkClient:            clerkClient,
		defaultStorageQuotaBytes: defaultStorageQuotaGB * 1024 * 1024 * 1024,
	}
}

// CreateUser handles the 'user.created' webhook from Clerk.
// It creates a new user with a "pending" status and sets initial Clerk metadata.
func (s *UserService) CreateUser(ctx context.Context, clerkID, email string) (*domain.User, error) {
	// 1. Create a new user in the local database with "pending" status.
	user := &domain.User{
		ID:                uuid.New(),
		ClerkUserID:       clerkID,
		Email:             email,
		Status:            "pending",              // Initial status
		Role:              "user",                 // Default role
		StorageQuotaBytes: s.defaultStorageQuotaBytes, // 5GB default quota
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create pending user: %w", err)
	}

	// 2. Update Clerk public metadata with the "pending" status and default role.
	metadata := map[string]interface{}{
		"status": "pending",
		"role":   "user",
	}
	if err := s.updateClerkPublicMetadata(ctx, clerkID, metadata); err != nil {
		// Log the error but don't fail the whole operation.
		// A retry mechanism or manual intervention might be needed here.
		log.Printf("ERROR: failed to set initial metadata for user %s: %v", clerkID, err)
	}

	return createdUser, nil
}

// ActivateUser activates a user's profile by updating their status and profile information.
// This is called via gRPC when the user submits their profile setup form.
func (s *UserService) ActivateUser(ctx context.Context, clerkID, fullName, username string) (*domain.User, error) {
	// 1. Get the user from the database.
	user, err := s.repo.GetByClerkID(ctx, clerkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by clerk id: %w", err)
	}

	// 2. Update the user's profile in the database.
	updates := map[string]interface{}{
		"full_name": fullName,
		"username":  username,
		"status":    "active",
	}
	updatedUser, err := s.repo.Update(ctx, user.ID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to activate user profile: %w", err)
	}

	// 3. Update Clerk public metadata with the "active" status.
	metadata := map[string]interface{}{
		"status": "active",
		"role":   user.Role, // Preserve the existing role
	}
	if err := s.updateClerkPublicMetadata(ctx, clerkID, metadata); err != nil {
		// Log the error but don't fail the whole operation.
		log.Printf("ERROR: failed to update metadata for activated user %s: %v", clerkID, err)
	}

	return updatedUser, nil
}

// GetUser retrieves a user by their Clerk ID from the context.
func (s *UserService) GetUser(ctx context.Context) (*domain.User, error) {
	clerkUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}
	return s.repo.GetByClerkID(ctx, clerkUserID)
}

// UpdateUser updates a user's profile information.
func (s *UserService) UpdateUser(ctx context.Context, email, fullName, username *string) (*domain.User, error) {
	clerkUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}

	userDomain, err := s.repo.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if email != nil {
		updates["email"] = *email
	}
	if fullName != nil {
		updates["full_name"] = *fullName
	}
	if username != nil {
		updates["username"] = *username
	}

	if len(updates) > 0 {
		// Also update on Clerk's side
		params := &user.UpdateParams{}
		if fullName != nil {
			params.FirstName = fullName
		}

		if _, err := user.Update(ctx, clerkUserID, params); err != nil {
			// Log the error but don't block the local update
			log.Printf("failed to update user on clerk: %v", err)
		}
	}

	if len(updates) == 0 {
		return userDomain, nil
	}

	return s.repo.Update(ctx, userDomain.ID, updates)
}

// DeleteUser soft deletes a user.
func (s *UserService) DeleteUser(ctx context.Context) error {
	clerkUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		return fmt.Errorf("user_id not found in context")
	}

	user, err := s.repo.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, user.ID)
}

// UpdateUserPreferences updates a user's preferences.
func (s *UserService) UpdateUserPreferences(ctx context.Context, theme, language, timezone *string, canvasSettings, notificationSettings, accessibilitySettings map[string]interface{}) (*domain.UserPreferences, error) {
	clerkUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}

	user, err := s.repo.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}

	prefs := &domain.UserPreferences{
		UserID: user.ID,
	}

	if theme != nil {
		prefs.Theme = *theme
	}
	if language != nil {
		prefs.Language = *language
	}
	if timezone != nil {
		prefs.Timezone = *timezone
	}
	if canvasSettings != nil {
		prefs.CanvasSettings = canvasSettings
	}
	if notificationSettings != nil {
		prefs.NotificationSettings = notificationSettings
	}
	if accessibilitySettings != nil {
		prefs.AccessibilitySettings = accessibilitySettings
	}

	return s.repo.UpdatePreferences(ctx, user.ID, prefs)
}

// CheckUserStatus checks if a user exists and if profile completion is needed
func (s *UserService) CheckUserStatus(ctx context.Context) (*domain.UserStatus, error) {
	clerkUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}

	user, err := s.repo.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		// User doesn't exist in our database but has a valid Clerk token
		return &domain.UserStatus{
			ProfileCompleted: false,
			NeedsRedirect:    true,
			RedirectURL:      "/profile-completion",
			User:             nil,
		}, nil
	}

	// Check if user profile is complete
	profileCompleted := s.isProfileComplete(user)

	status := &domain.UserStatus{
		ProfileCompleted: profileCompleted,
		NeedsRedirect:    !profileCompleted,
		RedirectURL:      "",
		User:             user,
	}

	if !profileCompleted {
		status.RedirectURL = "/profile-completion"
	}

	return status, nil
}

// isProfileComplete checks if the user profile has all required fields
func (s *UserService) isProfileComplete(user *domain.User) bool {
	return user.FullName != "" && user.Username != "" && user.Email != ""
}

// updateClerkPublicMetadata updates the user's public metadata in Clerk.
func (s *UserService) updateClerkPublicMetadata(ctx context.Context, clerkUserID string, metadata map[string]interface{}) error {
	if s.clerkClient == nil {
		return fmt.Errorf("clerk client not available")
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	rawMessage := json.RawMessage(metadataJSON)
	params := &user.UpdateParams{
		PublicMetadata: &rawMessage,
	}

	// Use the user package directly with the context
	_, err = user.Update(ctx, clerkUserID, params)
	if err != nil {
		return fmt.Errorf("failed to update user metadata in Clerk: %w", err)
	}

	log.Printf("Successfully updated public metadata for user %s in Clerk", clerkUserID)
	return nil
}
