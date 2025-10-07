package http

import (
	"encoding/json"
	"net/http"

	"github.com/raufhm/rra/internal/domain"
	"github.com/raufhm/rra/internal/middleware"
	"github.com/raufhm/rra/internal/usecase"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateSettingsRequest struct {
	Name            *string `json:"name"`
	CurrentPassword *string `json:"currentPassword"`
	NewPassword     *string `json:"newPassword"`
}

type CreateAPIKeyRequest struct {
	Name      string `json:"name"`
	ExpiresAt *string `json:"expiresAt"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Missing fields: email, password, name are required"})
		return
	}

	role := domain.UserRole(req.Role)
	if role == "" {
		role = domain.RoleUser
	}

	user, token, err := h.authUseCase.Register(req.Email, req.Password, req.Name, role)
	if err != nil {
		if err == usecase.ErrEmailExists {
			respondJSON(w, http.StatusConflict, map[string]string{"message": "Registration failed. Email may already be in use"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Registration failed"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
		"token": token,
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Missing fields: email and password are required"})
		return
	}

	user, token, err := h.authUseCase.Login(req.Email, req.Password)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Invalid email or password"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "An internal error occurred during login"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
		"token": token,
	})
}

// ForgotPassword handles password reset requests
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// UpdateUserSettings updates user settings
func (h *AuthHandler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	updatedUser, err := h.authUseCase.UpdateUserSettings(user.ID, req.Name, req.CurrentPassword, req.NewPassword)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    updatedUser.ID,
			"email": updatedUser.Email,
			"name":  updatedUser.Name,
			"role":  updatedUser.Role,
		},
		"message": "Settings updated successfully",
	})
}

// GetAPIKeys retrieves all API keys for the current user
func (h *AuthHandler) GetAPIKeys(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	keys, err := h.authUseCase.GetAPIKeys(user.ID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve API keys"})
		return
	}

	// Mask keys and add prefix for display
	maskedKeys := make([]map[string]interface{}, len(keys))
	for i, key := range keys {
		maskedKeys[i] = map[string]interface{}{
			"id":           key.ID,
			"name":         key.Name,
			"key_prefix":   "rr_live_***",
			"expires_at":   key.ExpiresAt,
			"last_used_at": key.LastUsedAt,
			"created_at":   key.CreatedAt,
			"active":       key.Active,
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"keys": maskedKeys})
}

// CreateAPIKey generates a new API key
func (h *AuthHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Key name is required"})
		return
	}

	rawKey, keyID, err := h.authUseCase.CreateAPIKey(user.ID, req.Name)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to generate API key"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"key": rawKey,
		"id":  keyID,
	})
}

// RevokeAPIKey revokes an API key
func (h *AuthHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	keyID := getIDFromPath(r, "/api/v1/auth/api-keys/")
	if keyID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid key ID"})
		return
	}

	if err := h.authUseCase.RevokeAPIKey(user.ID, keyID); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "API Key revoked successfully"})
}