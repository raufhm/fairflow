package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/raufhm/fairflow/services/member/internal/usecase"
	"github.com/raufhm/fairflow/shared/middleware"
)

type MemberHandler struct {
	memberUseCase *usecase.MemberUseCase
}

func NewMemberHandler(memberUseCase *usecase.MemberUseCase) *MemberHandler {
	return &MemberHandler{
		memberUseCase: memberUseCase,
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
	ctx := r.Context()
	groupID := getIDFromPath(r, "/api/v1/groups/", "/members")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	members, err := h.memberUseCase.GetMembers(ctx, groupID)
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

	ctx := r.Context()
	groupID := getIDFromPath(r, "/api/v1/groups/", "/members")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
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

	member, err := h.memberUseCase.CreateMember(ctx, groupID, user.ID, user.Name, req.Name, req.Email, req.Weight)
	if err != nil {
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

// GetMember retrieves a specific member
func (h *MemberHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	memberID := getIDFromPath(r, "/api/v1/members/")
	if memberID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid member ID"})
		return
	}

	member, err := h.memberUseCase.GetMember(ctx, memberID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve member"})
		return
	}
	if member == nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": "Member not found"})
		return
	}

	respondJSON(w, http.StatusOK, member)
}

// UpdateMember updates a member
func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	ctx := r.Context()
	memberID := getIDFromPath(r, "/api/v1/members/")
	if memberID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid member ID"})
		return
	}

	var req UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	if err := h.memberUseCase.UpdateMember(ctx, memberID, user.ID, user.Name, req.Name, req.Email, req.Weight, req.Active); err != nil {
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

	ctx := r.Context()
	memberID := getIDFromPath(r, "/api/v1/members/")
	if memberID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid member ID"})
		return
	}

	if err := h.memberUseCase.DeleteMember(ctx, memberID, user.ID, user.Name); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to delete member"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member deleted successfully"})
}

// GetMemberCapacity retrieves the capacity status of a member
func (h *MemberHandler) GetMemberCapacity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	memberID := getIDFromPath(r, "/api/v1/members/", "/capacity")
	if memberID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid member ID"})
		return
	}

	capacity, err := h.memberUseCase.GetMemberCapacity(ctx, memberID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, capacity)
}

// Helper functions

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
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
