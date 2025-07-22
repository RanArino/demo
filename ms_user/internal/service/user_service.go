package service

import (
	"context"
	"demo/ms_user/internal/domain"
	"fmt"

	"github.com/clerk/clerk-sdk-go/v2/user"
)

// UserService provides user-related business logic.
type UserService struct {
	repo domain.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser is deprecated in favor of the user.created webhook handler.
func (s *UserService) CreateUser(ctx context.Context, clerkID, email, fullName, username string) (*domain.User, error) {
	return nil, fmt.Errorf("CreateUser is deprecated")
}

// GetUser retrieves a user by their Clerk ID from the context.
func (s *UserService) GetUser(ctx context.Context) (*domain.User, error) {
	clerkUserID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}
	return s.repo.GetByClerkID(ctx, clerkUserID)
}

// UpdateUser updates a user's profile information.
func (s *UserService) UpdateUser(ctx context.Context, email, fullName, username *string) (*domain.User, error) {
	clerkUserID, ok := ctx.Value("user_id").(string)
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
			fmt.Printf("failed to update user on clerk: %v\n", err)
		}
	}

	if len(updates) == 0 {
		return userDomain, nil
	}

	return s.repo.Update(ctx, userDomain.ID, updates)
}

// DeleteUser soft deletes a user.
func (s *UserService) DeleteUser(ctx context.Context) error {
	clerkUserID, ok := ctx.Value("user_id").(string)
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
	clerkUserID, ok := ctx.Value("user_id").(string)
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
