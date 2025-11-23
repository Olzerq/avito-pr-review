package http

import (
	"context"
	"encoding/json"
	"github.com/Olzerq/avito-pr-reviewer/internal/service"
	"net/http"
	"time"
)

type StatsHandler struct {
	statsService *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// GET /stats/assignments
func (h *StatsHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, err := h.statsService.GetAssignmentsStats(ctx)
	if err != nil {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}
	
	byUser := make([]map[string]any, 0, len(stats.ByUser))
	for _, s := range stats.ByUser {
		byUser = append(byUser, map[string]any{
			"user_id":        s.UserID,
			"assigned_count": s.AssignedCount,
		})
	}

	byPR := make([]map[string]any, 0, len(stats.ByPR))
	for _, s := range stats.ByPR {
		byPR = append(byPR, map[string]any{
			"pull_request_id": s.PullRequestID,
			"reviewers_count": s.ReviewersCount,
		})
	}

	resp := map[string]any{
		"by_user": byUser,
		"by_pr":   byPR,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
