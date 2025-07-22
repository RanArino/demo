
package service

import (
	"context"
	"demo/ms_user/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/svix/svix-webhooks/go"
)

type WebhookService struct {
	userRepo domain.UserRepository
	wh       *svix.Webhook
}

func NewWebhookService(userRepo domain.UserRepository, secret string) (*WebhookService, error) {
	wh, err := svix.NewWebhook(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return &WebhookService{userRepo: userRepo, wh: wh}, nil
}

func (s *WebhookService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("webhook: failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := s.wh.Verify(body, r.Header); err != nil {
		log.Printf("webhook: verification failed: %v", err)
		http.Error(w, "Webhook verification failed", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	var evt struct {
		Type string `json:"type"`
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
		user := &domain.User{
			ClerkUserID: userCreatedEvent.ID,
			Email:       userCreatedEvent.Email,
			FullName:    userCreatedEvent.FullName,
			Username:    userCreatedEvent.Username,
		}
		_, err := s.userRepo.Create(ctx, user)
		if err != nil {
			log.Printf("webhook: failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		log.Printf("webhook: user.created event processed for user ID: %s", userCreatedEvent.ID)

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
			log.Printf("webhook: failed to find user for update: %v", err)
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
			log.Printf("webhook: failed to update user: %v", err)
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}
		log.Printf("webhook: user.updated event processed for user ID: %s", userUpdatedEvent.ID)

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
			log.Printf("webhook: failed to find user for deletion: %v", err)
			http.Error(w, "Failed to find user", http.StatusNotFound)
			return
		}
		if err := s.userRepo.Delete(ctx, user.ID); err != nil {
			log.Printf("webhook: failed to delete user: %v", err)
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
		log.Printf("webhook: user.deleted event processed for user ID: %s", userDeletedEvent.ID)

	default:
		log.Printf("webhook: unhandled event type: %s", evt.Type)
		http.Error(w, "Unhandled webhook event type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
