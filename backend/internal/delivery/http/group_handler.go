package http

import (
	"encoding/json"
	"net/http"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/middleware"
	"github.com/raufhm/fairflow/internal/usecase"
)

type GroupHandler struct {
	groupUseCase *usecase.GroupUseCase
}

func NewGroupHandler(groupUseCase *usecase.GroupUseCase) *GroupHandler {
	return &GroupHandler{groupUseCase: groupUseCase}
}

type CreateGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Strategy    string  `json:"strategy"`
}

type UpdateGroupRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
}

// GetAllGroups retrieves all groups
func (h *GroupHandler) GetAllGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.groupUseCase.GetAllGroups()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve groups"})
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

// CreateGroup creates a new group
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Group name is required"})
		return
	}

	strategy := domain.AssignmentStrategy(req.Strategy)
	if strategy == "" {
		strategy = domain.StrategyWeightedRoundRobin
	}

	group, err := h.groupUseCase.CreateGroup(user.ID, user.Name, req.Name, req.Description, strategy)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create group"})
		return
	}

	respondJSON(w, http.StatusCreated, group)
}

// GetGroup retrieves a specific group
func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r, "/api/v1/groups/")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	group, err := h.groupUseCase.GetGroup(id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve group"})
		return
	}
	if group == nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": "Group not found"})
		return
	}

	respondJSON(w, http.StatusOK, group)
}

// UpdateGroup updates a group
func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	id := getIDFromPath(r, "/api/v1/groups/")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	// Check if user can modify group
	canModify, err := h.groupUseCase.CanModifyGroup(id, user.ID, user.Role)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": "Group not found"})
		return
	}
	if !canModify {
		respondJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden: You do not have permission to modify this group"})
		return
	}

	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	group, err := h.groupUseCase.UpdateGroup(id, user.ID, user.Name, req.Name, req.Description, req.Active)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, group)
}

// DeleteGroup deletes a group
func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	id := getIDFromPath(r, "/api/v1/groups/")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	// Check if user can modify group
	canModify, err := h.groupUseCase.CanModifyGroup(id, user.ID, user.Role)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": "Group not found"})
		return
	}
	if !canModify {
		respondJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden: You do not have permission to modify this group"})
		return
	}

	if err := h.groupUseCase.DeleteGroup(id, user.ID, user.Name); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to delete group"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Group deleted successfully"})
}

type PauseGroupRequest struct {
	Reason *string `json:"reason"`
}

// PauseGroup pauses assignments for a group
func (h *GroupHandler) PauseGroup(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	id := getIDFromPath(r, "/api/v1/groups/", "/pause")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	// Check if user can modify group
	canModify, err := h.groupUseCase.CanModifyGroup(id, user.ID, user.Role)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": "Group not found"})
		return
	}
	if !canModify {
		respondJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden: You do not have permission to modify this group"})
		return
	}

	var req PauseGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is empty, that's okay, just use nil reason
		req.Reason = nil
	}

	if err := h.groupUseCase.PauseGroup(id, user.ID, user.Name, req.Reason); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Group assignments paused successfully"})
}

// ResumeGroup resumes assignments for a group
func (h *GroupHandler) ResumeGroup(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	id := getIDFromPath(r, "/api/v1/groups/", "/resume")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	// Check if user can modify group
	canModify, err := h.groupUseCase.CanModifyGroup(id, user.ID, user.Role)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": "Group not found"})
		return
	}
	if !canModify {
		respondJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden: You do not have permission to modify this group"})
		return
	}

	if err := h.groupUseCase.ResumeGroup(id, user.ID, user.Name); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Group assignments resumed successfully"})
}
