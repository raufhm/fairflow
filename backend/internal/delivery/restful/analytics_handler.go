package restful

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/raufhm/fairflow/internal/middleware"
	"github.com/raufhm/fairflow/internal/usecase"
)

type AnalyticsHandler struct {
	assignmentUseCase *usecase.AssignmentUseCase
	memberUseCase     *usecase.MemberUseCase
	groupUseCase      *usecase.GroupUseCase
}

func NewAnalyticsHandler(
	assignmentUseCase *usecase.AssignmentUseCase,
	memberUseCase *usecase.MemberUseCase,
	groupUseCase *usecase.GroupUseCase,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		assignmentUseCase: assignmentUseCase,
		memberUseCase:     memberUseCase,
		groupUseCase:      groupUseCase,
	}
}

// GetFairnessMetrics returns fairness metrics for a group
func (h *AnalyticsHandler) GetFairnessMetrics(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	groupID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, `{"message":"Invalid group ID"}`, http.StatusBadRequest)
		return
	}

	// Get all members with assignment counts
	members, err := h.memberUseCase.GetMembers(ctx, groupID)
	if err != nil {
		http.Error(w, `{"message":"Failed to fetch members"}`, http.StatusInternalServerError)
		return
	}

	// Calculate fairness metrics
	var totalAssignments int
	var maxAssignments int
	var minAssignments int = 999999
	activeMembers := 0

	for _, member := range members {
		if member.Active {
			activeMembers++
			totalAssignments += member.Assignments
			if member.Assignments > maxAssignments {
				maxAssignments = member.Assignments
			}
			if member.Assignments < minAssignments {
				minAssignments = member.Assignments
			}
		}
	}

	var avgAssignments float64
	var fairnessScore float64 = 100.0

	if activeMembers > 0 {
		avgAssignments = float64(totalAssignments) / float64(activeMembers)

		// Calculate fairness score (100 = perfect fairness, lower = less fair)
		if avgAssignments > 0 {
			deviation := float64(maxAssignments-minAssignments) / avgAssignments
			fairnessScore = 100.0 - (deviation * 10)
			if fairnessScore < 0 {
				fairnessScore = 0
			}
		}
	}

	metrics := map[string]interface{}{
		"group_id":          groupID,
		"total_assignments": totalAssignments,
		"active_members":    activeMembers,
		"avg_assignments":   avgAssignments,
		"max_assignments":   maxAssignments,
		"min_assignments":   minAssignments,
		"fairness_score":    fairnessScore,
		"variance":          maxAssignments - minAssignments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetWorkloadTrends returns workload trends over time
func (h *AnalyticsHandler) GetWorkloadTrends(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	groupID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, `{"message":"Invalid group ID"}`, http.StatusBadRequest)
		return
	}

	// Get assignments for the last 30 days (get all with high limit)
	assignments, _, err := h.assignmentUseCase.GetAssignments(ctx, groupID, 10000, 0)
	if err != nil {
		http.Error(w, `{"message":"Failed to fetch assignments"}`, http.StatusInternalServerError)
		return
	}

	// Group by date
	dailyCounts := make(map[string]int)
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	for _, assignment := range assignments {
		if assignment.CreatedAt.After(thirtyDaysAgo) {
			dateKey := assignment.CreatedAt.Format("2006-01-02")
			dailyCounts[dateKey]++
		}
	}

	// Convert to array of trend points
	var trends []map[string]interface{}
	for date, count := range dailyCounts {
		trends = append(trends, map[string]interface{}{
			"date":  date,
			"count": count,
		})
	}

	response := map[string]interface{}{
		"group_id": groupID,
		"period":   "30_days",
		"trends":   trends,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMemberPerformance returns performance metrics for all members
func (h *AnalyticsHandler) GetMemberPerformance(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	groupID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, `{"message":"Invalid group ID"}`, http.StatusBadRequest)
		return
	}

	members, err := h.memberUseCase.GetMembers(ctx, groupID)
	if err != nil {
		http.Error(w, `{"message":"Failed to fetch members"}`, http.StatusInternalServerError)
		return
	}

	var performance []map[string]interface{}
	for _, member := range members {
		perf := map[string]interface{}{
			"member_id":   member.ID,
			"name":        member.Name,
			"assignments": member.Assignments,
			"weight":      member.Weight,
			"active":      member.Active,
			"available":   member.Available,
		}
		performance = append(performance, perf)
	}

	response := map[string]interface{}{
		"group_id":    groupID,
		"performance": performance,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
