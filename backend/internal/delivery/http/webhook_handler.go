package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/raufhm/fairflow/internal/middleware"
	"github.com/raufhm/fairflow/internal/usecase"
)

type WebhookHandler struct {
	webhookUseCase *usecase.WebhookUseCase
}

func NewWebhookHandler(webhookUseCase *usecase.WebhookUseCase) *WebhookHandler {
	return &WebhookHandler{webhookUseCase: webhookUseCase}
}

// CreateWebhook creates a new webhook
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	groupID, err := strconv.ParseInt(chi.URLParam(r, "groupId"), 10, 64)
	if err != nil {
		http.Error(w, `{"message":"Invalid group ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	webhook, err := h.webhookUseCase.CreateWebhook(user.ID, groupID, req.URL, req.Events)
	if err != nil {
		http.Error(w, `{"message":"Failed to create webhook"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(webhook)
}

// GetWebhooks returns all webhooks for a group
func (h *WebhookHandler) GetWebhooks(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	groupID, err := strconv.ParseInt(chi.URLParam(r, "groupId"), 10, 64)
	if err != nil {
		http.Error(w, `{"message":"Invalid group ID"}`, http.StatusBadRequest)
		return
	}

	webhooks, err := h.webhookUseCase.GetWebhooksByGroup(groupID)
	if err != nil {
		http.Error(w, `{"message":"Failed to fetch webhooks"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}

// DeleteWebhook deletes a webhook
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	webhookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, `{"message":"Invalid webhook ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.webhookUseCase.DeleteWebhook(user.ID, webhookID); err != nil {
		http.Error(w, `{"message":"Failed to delete webhook"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
