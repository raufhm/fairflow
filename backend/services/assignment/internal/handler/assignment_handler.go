package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/raufhm/fairflow/services/assignment/internal/usecase"
	"github.com/raufhm/fairflow/shared/middleware"
)

type AssignmentHandler struct {
	assignmentUseCase *usecase.AssignmentUseCase
}

func NewAssignmentHandler(assignmentUseCase *usecase.AssignmentUseCase) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentUseCase: assignmentUseCase,
	}
}

type RecordAssignmentRequest struct {
	MemberID *int64  `json:"memberId"`
	Metadata *string `json:"metadata"`
}

// GetNextAssignee calculates the next assignee using weighted round-robin
func (h *AssignmentHandler) GetNextAssignee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	groupID := getIDFromPath(r, "/api/v1/groups/", "/next")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	member, err := h.assignmentUseCase.CalculateNextAssignee(ctx, groupID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"member":    member,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// RecordAssignment creates a new assignment record
func (h *AssignmentHandler) RecordAssignment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	ctx := r.Context()
	groupID := getIDFromPath(r, "/api/v1/groups/", "/assign")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	var req RecordAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	member, assignmentID, err := h.assignmentUseCase.RecordAssignment(ctx, groupID, user.ID, user.Name, req.MemberID, req.Metadata)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"assignmentId": assignmentID,
		"member":       member,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// GetAssignments retrieves assignment history for a group with pagination
func (h *AssignmentHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	groupID := getIDFromPath(r, "/api/v1/groups/", "/assignments")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	assignments, total, err := h.assignmentUseCase.GetAssignments(ctx, groupID, limit, offset)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve assignments"})
		return
	}

	page := (offset / limit) + 1
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"assignments": assignments,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// GetStats retrieves assignment statistics and distribution for a group
func (h *AssignmentHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	groupID := getIDFromPath(r, "/api/v1/groups/", "/stats")
	if groupID == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	stats, err := h.assignmentUseCase.GetStats(ctx, groupID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve statistics"})
		return
	}

	respondJSON(w, http.StatusOK, stats)
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
