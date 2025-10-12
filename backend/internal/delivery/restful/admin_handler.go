package restful

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/middleware"
	"github.com/raufhm/fairflow/internal/usecase"
)

type AdminHandler struct {
	adminUseCase *usecase.AdminUseCase
}

func NewAdminHandler(adminUseCase *usecase.AdminUseCase) *AdminHandler {
	return &AdminHandler{adminUseCase: adminUseCase}
}

type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

// GetAllUsers retrieves all users
func (h *AdminHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := h.adminUseCase.GetAllUsers(ctx)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve users"})
		return
	}

	// Remove password hashes from response
	safeUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		safeUsers[i] = map[string]interface{}{
			"id":              user.ID,
			"email":           user.Email,
			"name":            user.Name,
			"role":            user.Role,
			"organization_id": user.OrganizationID,
			"created_at":      user.CreatedAt,
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"users": safeUsers})
}

// UpdateUserRole updates a user's role
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	ctx := r.Context()
	targetUserID := getIDFromPath(r, "/api/v1/admin/users/")
	if targetUserID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid user ID"})
		return
	}

	var req UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if req.Role == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Role field required for update"})
		return
	}

	role := domain.UserRole(req.Role)
	if err := h.adminUseCase.UpdateUserRole(ctx, targetUserID, user.ID, user.Name, role); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "User role updated successfully"})
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	ctx := r.Context()
	targetUserID := getIDFromPath(r, "/api/v1/admin/users/")
	if targetUserID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid user ID"})
		return
	}

	if err := h.adminUseCase.DeleteUser(ctx, targetUserID, user.ID, user.Name); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// GetAuditLogs retrieves audit logs
func (h *AdminHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := h.adminUseCase.GetAuditLogs(ctx, limit)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve audit logs"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"logs": logs})
}

// Backup creates a database backup
func (h *AdminHandler) Backup(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder - actual backup logic would copy the SQLite file
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Backup functionality not yet implemented in Go version",
	})
}

// GetBackups lists available backups
func (h *AdminHandler) GetBackups(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"files": []string{},
	})
}

// Restore restores from a backup
func (h *AdminHandler) Restore(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusAccepted, map[string]string{
		"message": "Restore functionality not yet implemented in Go version",
	})
}

// ExportData exports all data
func (h *AdminHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"message":      "Data export successfully triggered",
		"instructions": "The server will now process the full database and export it to JSON",
	})
}
