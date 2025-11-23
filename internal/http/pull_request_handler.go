package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Olzerq/avito-pr-reviewer/internal/repository"
	"github.com/Olzerq/avito-pr-reviewer/internal/service"
	"net/http"
	"time"
)

type PullRequestHandler struct {
	prService *service.PullRequestService
}

func NewPullRequestHandler(prService *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{prService: prService}
}

type prCreateRequest struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type prReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"` // тут почему-то в example в openapi стоит другое название :(
}

type prMergeRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// POST /pullRequest/create
func (h *PullRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req prCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.ID == "" || req.Name == "" || req.AuthorID == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	pr, err := h.prService.Create(ctx, req.ID, req.Name, req.AuthorID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound),
			errors.Is(err, repository.ErrTeamNotFound):
			// 404 Автор/команда не найдены
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "NOT_FOUND",
					Message: "resource not found",
				},
			})
			return
		case errors.Is(err, repository.ErrPRExists):
			// 409 PR_EXISTS
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "PR_EXISTS",
					Message: "PR id already exists",
				},
			})
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	assigned := make([]string, 0, len(pr.Reviewers))
	for _, u := range pr.Reviewers {
		assigned = append(assigned, u.ID)
	}
	resp := map[string]any{
		"pr": map[string]any{
			"pull_request_id":    pr.ID,
			"pull_request_name":  pr.Name,
			"author_id":          pr.AuthorID,
			"status":             pr.Status,
			"assigned_reviewers": assigned,
			"createdAt":          nil, // эту колонку еще не сделал
			"mergedAt":           nil,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PullRequestHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req prReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.PullRequestID == "" || req.OldUserID == "" {
		http.Error(w, "pull_request_id and old_reviewer_id are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	pr, replacedBy, err := h.prService.Reassign(ctx, req.PullRequestID, req.OldUserID)
	if err != nil {
		switch {
		// 404
		case errors.Is(err, repository.ErrPRNotFound),
			errors.Is(err, repository.ErrUserNotFound):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "NOT_FOUND",
					Message: "resource not found",
				},
			})
			return

		// 409
		case errors.Is(err, service.ErrPRMerged):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "PR_MERGED",
					Message: "cannot reassign on merged PR",
				},
			})
			return

		// 409
		case errors.Is(err, repository.ErrReviewerNotAssigned):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "NOT_ASSIGNED",
					Message: "reviewer is not assigned to this PR",
				},
			})
			return

		// 409
		case errors.Is(err, service.ErrNoCandidate):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "NO_CANDIDATE",
					Message: "no active replacement candidate in team",
				},
			})
			return

		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	assigned := make([]string, 0, len(pr.Reviewers))
	for _, u := range pr.Reviewers {
		assigned = append(assigned, u.ID)
	}

	resp := map[string]any{
		"pr": map[string]any{
			"pull_request_id":    pr.ID,
			"pull_request_name":  pr.Name,
			"author_id":          pr.AuthorID,
			"status":             pr.Status,
			"assigned_reviewers": assigned,
			"createdAt":          nil,
			"mergedAt":           pr.MergedAt,
		},
		"replaced_by": replacedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /pullRequest/merge
func (h *PullRequestHandler) Merge(w http.ResponseWriter, r *http.Request) {
	var req prMergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.PullRequestID == "" {
		http.Error(w, "pull_request_id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	pr, err := h.prService.Merge(ctx, req.PullRequestID)
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "NOT_FOUND",
					Message: "resource not found",
				},
			})
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	assigned := make([]string, 0, len(pr.Reviewers))
	for _, u := range pr.Reviewers {
		assigned = append(assigned, u.ID)
	}

	resp := map[string]any{
		"pr": map[string]any{
			"pull_request_id":    pr.ID,
			"pull_request_name":  pr.Name,
			"author_id":          pr.AuthorID,
			"status":             pr.Status,
			"assigned_reviewers": assigned,
			"createdAt":          nil,
			"mergedAt":           pr.MergedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GET /users/getReview
func (h *PullRequestHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	prs, err := h.prService.GetUserReviews(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "NOT_FOUND",
					Message: "resource not found",
				},
			})
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respPRs := make([]map[string]any, 0, len(prs))
	for _, pr := range prs {
		respPRs = append(respPRs, map[string]any{
			"pull_request_id":   pr.ID,
			"pull_request_name": pr.Name,
			"author_id":         pr.AuthorID,
			"status":            pr.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id":       userID,
		"pull_requests": respPRs,
	})
}
