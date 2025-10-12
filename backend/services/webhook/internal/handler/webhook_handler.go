package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/raufhm/fairflow/services/webhook/internal/usecase"
	"github.com/raufhm/fairflow/shared/middleware"
)

type WebhookHandler struct {
	webhookUseCase *usecase.WebhookUseCase
}

func NewWebhookHandler(webhookUseCase *usecase.WebhookUseCase) *WebhookHandler {
	return &WebhookHandler{webhookUseCase: webhookUseCase}
}

type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// CreateWebhook creates a new webhook
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		return
	}

	ctx := r.Context()
	groupID := getIDFromPath(r, "/api/v1/groups/", "/webhooks")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if req.URL == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "URL is required"})
		return
	}

	webhook, err := h.webhookUseCase.CreateWebhook(ctx, user.ID, groupID, user.Name, req.URL, req.Events)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create webhook"})
		return
	}

	respondJSON(w, http.StatusCreated, webhook)
}

// GetWebhooks returns all webhooks for a group
func (h *WebhookHandler) GetWebhooks(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		return
	}

	ctx := r.Context()
	groupID := getIDFromPath(r, "/api/v1/groups/", "/webhooks")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	webhooks, err := h.webhookUseCase.GetWebhooksByGroup(ctx, groupID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch webhooks"})
		return
	}

	respondJSON(w, http.StatusOK, webhooks)
}

// DeleteWebhook deletes a webhook
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		return
	}

	ctx := r.Context()
	webhookID := getIDFromPath(r, "/api/v1/webhooks/")
	if webhookID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid webhook ID"})
		return
	}

	if err := h.webhookUseCase.DeleteWebhook(ctx, user.ID, user.Name, webhookID); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to delete webhook"})
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

// Helper functions

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// getIDFromPath extracts an ID from the URL path
func getIDFromPath(r *http.Request, prefix string, suffixes ...string) int64 {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	for _, suffix := range suffixes {
		path = strings.TrimSuffix(path, suffix)
	}
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parseID(parts[0])
	}
	return 0
}

// parseID converts a string to int64
func parseID(s string) int64 {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
