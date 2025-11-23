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

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type setIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type userResponse struct {
	User struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
	} `json:"user"`
}

// POST /user/setIsActive
func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req setIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.UserID == "" {
		http.Error(w, "user_id is empty", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.userService.SetIsActive(ctx, req.UserID, req.IsActive); err != nil {
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
		http.Error(w, "failed to set is_active", http.StatusInternalServerError)
		return
	}

	user, err := h.userService.GetByID(ctx, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp userResponse
	resp.User.UserID = user.ID
	resp.User.Username = user.Username
	resp.User.TeamName = user.TeamName
	resp.User.IsActive = user.IsActive

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
