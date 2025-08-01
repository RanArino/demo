package service

import (
	"context"
	"demo/ms_user/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	svix "github.com/svix/svix-webhooks/go"
)

type WebhookService struct {
	userRepo    domain.UserRepository
	wh          *svix.Webhook
	maxBodySize int64
}

func NewWebhookService(userRepo domain.UserRepository, secret string, maxBodySize int64) (*WebhookService, error) {
	wh, err := svix.NewWebhook(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return &WebhookService{userRepo: userRepo, wh: wh, maxBodySize: maxBodySize}, nil
}

func (s *WebhookService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Enhanced security: Check request method
	if r.Method != http.MethodPost {
		log.Printf("webhook: invalid method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Enhanced security: Check Content-Type header
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		log.Printf("webhook: invalid content type: %s", contentType)
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	// Enhanced security: Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, s.maxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("webhook: failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Enhanced security: Validate minimum body size
	if len(body) == 0 {
		log.Printf("webhook: empty request body")
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	if err := s.wh.Verify(body, r.Header); err != nil {
		log.Printf("webhook: verification failed: %v", err)
		http.Error(w, "Webhook verification failed", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	var evt struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &evt); err != nil {
		log.Printf("webhook: failed to unmarshal event: %v", err)
		http.Error(w, "Failed to unmarshal event", http.StatusBadRequest)
		return
	}

	switch evt.Type {
	case "user.created":
		var userCreatedEvent struct {
			ID       string `json:"id"`
			Email    string `json:"email_address"`
			FullName string `json:"full_name"`
			Username string `json:"username"`
		}
		if err := json.Unmarshal(evt.Data, &userCreatedEvent); err != nil {
			log.Printf("webhook: failed to unmarshal user.created event: %v", err)
			http.Error(w, "Failed to unmarshal user.created event", http.StatusBadRequest)
			return
		}

		// Enhanced security: Validate required fields and format
		if err := s.validateUserCreatedEvent(userCreatedEvent); err != nil {
			log.Printf("webhook: user.created validation failed: %v", err)
			http.Error(w, "Invalid user data", http.StatusBadRequest)
			return
		}

		user := &domain.User{
			ClerkUserID: userCreatedEvent.ID,
			Email:       userCreatedEvent.Email,
			FullName:    userCreatedEvent.FullName,
			Username:    userCreatedEvent.Username,
			Role:        "user", // Default role for new users
		}
		createdUser, err := s.userRepo.Create(ctx, user)
		if err != nil {
			log.Printf("webhook: failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// Audit log: User creation via webhook
		log.Printf("AUDIT: user created via webhook - clerk_id=%s, email=%s, username=%s, id=%s, timestamp=%s",
			userCreatedEvent.ID, userCreatedEvent.Email, userCreatedEvent.Username,
			createdUser.ID.String(), time.Now().UTC().Format(time.RFC3339))

		log.Printf("webhook: user.created event processed successfully")

	case "user.updated":
		var userUpdatedEvent struct {
			ID       string `json:"id"`
			Email    string `json:"email_address"`
			FullName string `json:"full_name"`
			Username string `json:"username"`
		}
		if err := json.Unmarshal(evt.Data, &userUpdatedEvent); err != nil {
			log.Printf("webhook: failed to unmarshal user.updated event: %v", err)
			http.Error(w, "Failed to unmarshal user.updated event", http.StatusBadRequest)
			return
		}
		user, err := s.userRepo.GetByClerkID(ctx, userUpdatedEvent.ID)
		if err != nil {
			log.Printf("webhook: failed to find user for update")
			http.Error(w, "Failed to find user", http.StatusNotFound)
			return
		}
		updates := map[string]interface{}{
			"email":     userUpdatedEvent.Email,
			"full_name": userUpdatedEvent.FullName,
			"username":  userUpdatedEvent.Username,
		}
		_, err = s.userRepo.Update(ctx, user.ID, updates)
		if err != nil {
			log.Printf("webhook: failed to update user")
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}
		log.Printf("webhook: user.updated event processed")

	case "user.deleted":
		var userDeletedEvent struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(evt.Data, &userDeletedEvent); err != nil {
			log.Printf("webhook: failed to unmarshal user.deleted event: %v", err)
			http.Error(w, "Failed to unmarshal user.deleted event", http.StatusBadRequest)
			return
		}
		user, err := s.userRepo.GetByClerkID(ctx, userDeletedEvent.ID)
		if err != nil {
			log.Printf("webhook: failed to find user for deletion")
			http.Error(w, "Failed to find user", http.StatusNotFound)
			return
		}
		if err := s.userRepo.Delete(ctx, user.ID); err != nil {
			log.Printf("webhook: failed to delete user")
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
		log.Printf("webhook: user.deleted event processed")

	default:
		log.Printf("webhook: unhandled event type: %s", evt.Type)
		http.Error(w, "Unhandled webhook event type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// validateUserCreatedEvent validates the user.created event data
func (s *WebhookService) validateUserCreatedEvent(event struct {
	ID       string `json:"id"`
	Email    string `json:"email_address"`
	FullName string `json:"full_name"`
	Username string `json:"username"`
}) error {
	// Validate Clerk User ID format (starts with "user_")
	if event.ID == "" || !strings.HasPrefix(event.ID, "user_") {
		return fmt.Errorf("invalid clerk user ID format")
	}

	// Validate email format (RFC 5322)
	if _, err := mail.ParseAddress(event.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}

	// Validate field lengths to prevent database overflow
	if len(event.ID) > 100 || len(event.Email) > 255 ||
		len(event.FullName) > 255 || len(event.Username) > 100 {
		return fmt.Errorf("field length exceeds maximum allowed")
	}

	// Validate username (alphanumeric plus common characters)
	if event.Username != "" {
		for _, char := range event.Username {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '_' || char == '-' || char == '.') {
				return fmt.Errorf("username contains invalid characters")
			}
		}
	}

	return nil
}
