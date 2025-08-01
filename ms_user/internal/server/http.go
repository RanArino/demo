package server

import (
	"demo/ms_user/internal/service"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/svix/svix-webhooks/go"
)

// WebhookHandler is a struct that holds the dependencies for the webhook handler.
type WebhookHandler struct {
	userService   *service.UserService
	webhookSecret string
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(userService *service.UserService, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		userService:   userService,
		webhookSecret: webhookSecret,
	}
}

// ServeHTTP is the entry point for the webhook handler.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("INFO: Webhook request received - Method: %s, URL: %s, Remote: %s", r.Method, r.URL.String(), r.RemoteAddr)
	log.Printf("INFO: Webhook headers: %v", r.Header)
	
	if r.Method != http.MethodPost {
		log.Printf("ERROR: Invalid method %s, expected POST", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read webhook body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	
	log.Printf("INFO: Webhook body received - Length: %d bytes", len(body))
	log.Printf("DEBUG: Webhook body content: %s", string(body))

	// Verify the webhook signature
	log.Printf("INFO: Creating webhook verifier with secret (first 10 chars): %s...", h.webhookSecret[:10])
	wh, err := svix.NewWebhook(h.webhookSecret)
	if err != nil {
		log.Printf("ERROR: Failed to create webhook verifier: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Verifying webhook signature...")
	if err := wh.Verify(body, r.Header); err != nil {
		log.Printf("ERROR: Failed to verify webhook signature: %v", err)
		log.Printf("DEBUG: Available headers for verification: %v", r.Header)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("INFO: Webhook signature verified successfully")

	// Decode the webhook payload
	log.Printf("INFO: Parsing webhook payload...")
	var payload struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("ERROR: Failed to unmarshal webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	
	log.Printf("INFO: Webhook event type: %s", payload.Type)

	// Handle the event type
	if payload.Type == "user.created" {
		log.Printf("INFO: Processing user.created event")
		h.handleUserCreated(w, r, payload.Data)
	} else {
		log.Printf("INFO: Received unhandled webhook event type: %s - Acknowledging", payload.Type)
		w.WriteHeader(http.StatusOK) // Acknowledge other events
		json.NewEncoder(w).Encode(map[string]string{"status": "acknowledged", "type": payload.Type})
	}
}

func (h *WebhookHandler) handleUserCreated(w http.ResponseWriter, r *http.Request, data json.RawMessage) {
	log.Printf("INFO: Starting handleUserCreated with data: %s", string(data))
	
	var userData struct {
		ID         string `json:"id"`
		Email      string `json:"email_address"`
		// Add other fields from the Clerk user object as needed
	}
	// It's nested under email_addresses array in the actual payload
	var clerkUser struct {
		ID            string `json:"id"`
		EmailAddresses []struct {
			EmailAddress string `json:"email_address"`
			ID           string `json:"id"`
		} `json:"email_addresses"`
	}

	log.Printf("INFO: Parsing user.created data...")
	if err := json.Unmarshal(data, &clerkUser); err != nil {
		log.Printf("ERROR: Failed to unmarshal user.created data: %v", err)
		log.Printf("DEBUG: Raw data was: %s", string(data))
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}
	
	log.Printf("INFO: Parsed clerk user - ID: %s, EmailAddresses count: %d", clerkUser.ID, len(clerkUser.EmailAddresses))

	if len(clerkUser.EmailAddresses) == 0 {
		log.Printf("ERROR: No email address found for user %s in webhook", clerkUser.ID)
		http.Error(w, "No email address found", http.StatusBadRequest)
		return
	}
	
	userData.ID = clerkUser.ID
	userData.Email = clerkUser.EmailAddresses[0].EmailAddress
	
	log.Printf("INFO: Extracted user data - ID: %s, Email: %s", userData.ID, userData.Email)
	log.Printf("INFO: Calling userService.CreateUser...")

	createdUser, err := h.userService.CreateUser(r.Context(), userData.ID, userData.Email)
	if err != nil {
		log.Printf("ERROR: Failed to create user in service: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	log.Printf("SUCCESS: User created successfully - ID: %s, ClerkID: %s, Email: %s", createdUser.ID, createdUser.ClerkUserID, createdUser.Email)

	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"status": "success", "user_id": createdUser.ID.String(), "clerk_id": userData.ID}
	json.NewEncoder(w).Encode(response)
	log.Printf("INFO: Webhook response sent successfully")
}

// StartHTTPServer starts the HTTP server for webhooks.
func StartHTTPServer(port string, userService *service.UserService, webhookSecret string) {
	log.Printf("INFO: Initializing webhook server...")
	log.Printf("INFO: Webhook secret configured: %s... (length: %d)", webhookSecret[:10], len(webhookSecret))
	
	handler := NewWebhookHandler(userService, webhookSecret)
	mux := http.NewServeMux()
	mux.Handle("/api/v1/webhooks/clerk", handler)
	
	log.Printf("INFO: Webhook endpoint registered at: /api/v1/webhooks/clerk")
	log.Printf("INFO: HTTP webhook server listening on port %s", port)
	log.Printf("INFO: Full webhook URL will be: http://localhost:%s/api/v1/webhooks/clerk", port)
	
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("ERROR: Failed to start HTTP server: %v", err)
	}
}
