package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/raufhm/fairflow/internal/middleware"
	"github.com/raufhm/fairflow/internal/usecase"
)

type AssignmentHandler struct {
	assignmentUseCase *usecase.AssignmentUseCase
}

func NewAssignmentHandler(assignmentUseCase *usecase.AssignmentUseCase) *AssignmentHandler {
	return &AssignmentHandler{assignmentUseCase: assignmentUseCase}
}

type RecordAssignmentRequest struct {
	MemberID *int64  `json:"memberId"`
	Metadata *string `json:"metadata"`
}

// GetNextAssignee calculates the next assignee
func (h *AssignmentHandler) GetNextAssignee(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r, "/api/v1/groups/", "/next")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	member, err := h.assignmentUseCase.CalculateNextAssignee(id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"member":    member,
		"timestamp": getCurrentTimestamp(),
	})
}

// RecordAssignment creates a new assignment
func (h *AssignmentHandler) RecordAssignment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	id := getIDFromPath(r, "/api/v1/groups/", "/assign")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	var req RecordAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	member, assignmentID, err := h.assignmentUseCase.RecordAssignment(id, user.ID, req.MemberID, req.Metadata)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"assignmentId": assignmentID,
		"member":       member,
		"timestamp":    getCurrentTimestamp(),
	})
}

// GetAssignments retrieves assignment history for a group
func (h *AssignmentHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r, "/api/v1/groups/", "/assignments")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	assignments, total, err := h.assignmentUseCase.GetAssignments(id, limit, offset)
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

// GetStats retrieves assignment statistics for a group
func (h *AssignmentHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r, "/api/v1/groups/", "/stats")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid group ID"})
		return
	}

	stats, err := h.assignmentUseCase.GetStats(id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve statistics"})
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// CompleteAssignment marks an assignment as completed
func (h *AssignmentHandler) CompleteAssignment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	id := getIDFromPath(r, "/api/v1/assignments/", "/complete")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid assignment ID"})
		return
	}

	if err := h.assignmentUseCase.CompleteAssignment(id); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Assignment completed successfully",
		"timestamp": getCurrentTimestamp(),
	})
}

// CancelAssignment marks an assignment as cancelled
func (h *AssignmentHandler) CancelAssignment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
		return
	}

	id := getIDFromPath(r, "/api/v1/assignments/", "/cancel")
	if id == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid assignment ID"})
		return
	}

	if err := h.assignmentUseCase.CancelAssignment(id); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Assignment cancelled successfully",
		"timestamp": getCurrentTimestamp(),
	})
}
