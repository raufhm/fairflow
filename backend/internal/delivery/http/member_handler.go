package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/raufhm/rra/internal/middleware"
	"github.com/raufhm/rra/internal/usecase"
)

type MemberHandler struct {
	memberUseCase *usecase.MemberUseCase
	groupUseCase  *usecase.GroupUseCase
}

func NewMemberHandler(memberUseCase *usecase.MemberUseCase, groupUseCase *usecase.GroupUseCase) *MemberHandler {
	return &MemberHandler{
		memberUseCase: memberUseCase,
		groupUseCase:  groupUseCase,
	}
}

type CreateMemberRequest struct {
	Name   string  `json:"name"`
	Email  *string `json:"email"`
	Weight int     `json:"weight"`
}

type UpdateMemberRequest struct {
	Name   *string `json:"name"`
	Email  *string `json:"email"`
	Weight *int    `json:"weight"`
	Active *bool   `json:"active"`
}

// GetMembers retrieves all members of a group
func (h *MemberHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	groupID := getIDFromPath(r, "/api/v1/groups/", "/members")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	members, err := h.memberUseCase.GetMembers(groupID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve members"})
		return
	}

	respondJSON(w, http.StatusOK, members)
}

// CreateMember creates a new member in a group
func (h *MemberHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	groupID := getIDFromPath(r, "/api/v1/groups/", "/members")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	// Check if user can modify group
	canModify, err := h.groupUseCase.CanModifyGroup(groupID, user.ID, user.Role)
	if err != nil || !canModify {
		respondJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden: You do not have permission to modify this group"})
		return
	}

	var req CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Member name is required"})
		return
	}

	if req.Weight == 0 {
		req.Weight = 100
	}

	member, err := h.memberUseCase.CreateMember(groupID, user.ID, user.Name, req.Name, req.Email, req.Weight)
	if err != nil {
		// Log the actual error for debugging
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to add member",
			"error":   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":   member.ID,
		"name": member.Name,
	})
}

// UpdateMember updates a member
func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	groupID, memberID := getMemberIDs(r)
	if groupID == 0 || memberID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group or member ID"})
		return
	}

	// Check if user can modify group
	canModify, err := h.groupUseCase.CanModifyGroup(groupID, user.ID, user.Role)
	if err != nil || !canModify {
		respondJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden: You do not have permission to modify this group"})
		return
	}

	var req UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if err := h.memberUseCase.UpdateMember(memberID, user.ID, user.Name, req.Name, req.Email, req.Weight, req.Active); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member updated successfully"})
}

// DeleteMember deletes a member
func (h *MemberHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	groupID, memberID := getMemberIDs(r)
	if groupID == 0 || memberID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group or member ID"})
		return
	}

	// Check if user can modify group
	canModify, err := h.groupUseCase.CanModifyGroup(groupID, user.ID, user.Role)
	if err != nil || !canModify {
		respondJSON(w, http.StatusForbidden, map[string]string{"message": "Forbidden: You do not have permission to modify this group"})
		return
	}

	if err := h.memberUseCase.DeleteMember(memberID, user.ID, user.Name); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to delete member"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member deleted successfully"})
}

// GetMemberCapacity retrieves the capacity status of a member
func (h *MemberHandler) GetMemberCapacity(w http.ResponseWriter, r *http.Request) {
	memberID := getIDFromPath(r, "/api/v1/members/", "/capacity")
	if memberID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid member ID"})
		return
	}

	capacity, err := h.memberUseCase.GetMemberCapacity(memberID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, capacity)
}

// Helper function to get group and member IDs from path
func getMemberIDs(r *http.Request) (int64, int64) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/")
	parts := strings.Split(path, "/")
	if len(parts) >= 3 {
		groupID := parseID(parts[0])
		memberID := parseID(parts[2])
		return groupID, memberID
	}
	return 0, 0
}