package repository

import (
	"context"
	"demo/ms_user/ent"
	"demo/ms_user/ent/user"
	"demo/ms_user/ent/userpreferences"
	"demo/ms_user/internal/domain"
	"time"

	"github.com/google/uuid"
)

type entUserRepository struct {
	client *ent.Client
}

func NewEntUserRepository(client *ent.Client) domain.UserRepository {
	return &entUserRepository{client: client}
}

func (r *entUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	entUser, err := r.client.User.
		Create().
		SetID(user.ID).
		SetClerkUserID(user.ClerkUserID).
		SetEmail(user.Email).
		SetFullName(user.FullName).
		SetUsername(user.Username).
		SetRole(user.Role).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainUser(entUser), nil
}

func (r *entUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	entUser, err := r.client.User.
		Query().
		Where(user.ID(id)).
		WithPreferences(func(q *ent.UserPreferencesQuery) {
			q.WithUser()
		}).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainUser(entUser), nil
}

func (r *entUserRepository) GetByClerkID(ctx context.Context, clerkID string) (*domain.User, error) {
	entUser, err := r.client.User.
		Query().
		Where(user.ClerkUserID(clerkID)).
		WithPreferences(func(q *ent.UserPreferencesQuery) {
			q.WithUser()
		}).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainUser(entUser), nil
}

func (r *entUserRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*domain.User, error) {
	updater := r.client.User.UpdateOneID(id)
	// This is a bit manual, but it's safe and explicit.
	if val, ok := updates["email"]; ok {
		updater.SetEmail(val.(string))
	}
	if val, ok := updates["full_name"]; ok {
		updater.SetFullName(val.(string))
	}
	if val, ok := updates["username"]; ok {
		updater.SetUsername(val.(string))
	}
	if val, ok := updates["role"]; ok {
		updater.SetRole(val.(string))
	}

	_, err := updater.Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *entUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Ent doesn't have a direct soft-delete method in the generated code.
	// We implement it by setting the deleted_at field.
	_, err := r.client.User.UpdateOneID(id).SetDeletedAt(time.Now()).Save(ctx)
	return err
}

func (r *entUserRepository) UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.UserPreferences) (*domain.UserPreferences, error) {
	// "Upsert" logic for preferences
	prefID, err := r.client.UserPreferences.
		Query().
		Where(userpreferences.HasUserWith(user.ID(userID))).
		OnlyID(ctx)

	var entPrefs *ent.UserPreferences
	if ent.IsNotFound(err) {
		// Create new preferences
		entPrefs, err = r.client.UserPreferences.
			Create().
			SetUserID(userID).
			SetTheme(prefs.Theme).
			SetLanguage(prefs.Language).
			SetTimezone(prefs.Timezone).
			SetCanvasSettings(prefs.CanvasSettings).
			SetNotificationSettings(prefs.NotificationSettings).
			SetAccessibilitySettings(prefs.AccessibilitySettings).
			Save(ctx)
	} else if err == nil {
		// Update existing preferences
		updater := r.client.UserPreferences.UpdateOneID(prefID)
		if prefs.Theme != "" {
			updater.SetTheme(prefs.Theme)
		}
		if prefs.Language != "" {
			updater.SetLanguage(prefs.Language)
		}
		if prefs.Timezone != "" {
			updater.SetTimezone(prefs.Timezone)
		}
		if prefs.CanvasSettings != nil {
			updater.SetCanvasSettings(prefs.CanvasSettings)
		}
		if prefs.NotificationSettings != nil {
			updater.SetNotificationSettings(prefs.NotificationSettings)
		}
		if prefs.AccessibilitySettings != nil {
			updater.SetAccessibilitySettings(prefs.AccessibilitySettings)
		}
		entPrefs, err = updater.Save(ctx)
	}

	if err != nil {
		return nil, err
	}

	// Reload the entity to ensure all edges are loaded, especially the User edge.
	entPrefs, err = r.client.UserPreferences.
		Query().
		Where(userpreferences.ID(entPrefs.ID)).
		WithUser().
		Only(ctx)

	if err != nil {
		return nil, err
	}

	return toDomainUserPreferences(entPrefs), nil
}

// --- Conversion Helpers ---

func toDomainUser(entUser *ent.User) *domain.User {
	if entUser == nil {
		return nil
	}
	domainUser := &domain.User{
		ID:                entUser.ID,
		ClerkUserID:       entUser.ClerkUserID,
		Email:             entUser.Email,
		FullName:          entUser.FullName,
		Username:          entUser.Username,
		Role:              entUser.Role,
		StorageUsedBytes:  entUser.StorageUsedBytes,
		StorageQuotaBytes: entUser.StorageQuotaBytes,
		Status:            entUser.Status,
		CreatedAt:         entUser.CreatedAt,
		UpdatedAt:         entUser.UpdatedAt,
		DeletedAt:         entUser.DeletedAt,
	}

	if entUser.Edges.Preferences != nil {
		domainUser.Preferences = toDomainUserPreferences(entUser.Edges.Preferences)
	}

	return domainUser
}

func toDomainUserPreferences(entPrefs *ent.UserPreferences) *domain.UserPreferences {
	if entPrefs == nil {
		return nil
	}

	var userID uuid.UUID
	if entPrefs.Edges.User != nil {
		userID = entPrefs.Edges.User.ID
	}

	return &domain.UserPreferences{
		ID:                    entPrefs.ID,
		UserID:                userID,
		Theme:                 entPrefs.Theme,
		Language:              entPrefs.Language,
		Timezone:              entPrefs.Timezone,
		CanvasSettings:        entPrefs.CanvasSettings,
		NotificationSettings:  entPrefs.NotificationSettings,
		AccessibilitySettings: entPrefs.AccessibilitySettings,
		CreatedAt:             entPrefs.CreatedAt,
		UpdatedAt:             entPrefs.UpdatedAt,
	}
}