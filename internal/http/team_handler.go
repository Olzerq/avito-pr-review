package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Olzerq/avito-pr-reviewer/internal/model"
	"github.com/Olzerq/avito-pr-reviewer/internal/repository"
	"github.com/Olzerq/avito-pr-reviewer/internal/service"
	"net/http"
	"time"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// структура запроса для /team/add
type teamAddRequest struct {
	TeamName string `json:"team_name"`
	Members  []struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	} `json:"members"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// /team/add

func (h *TeamHandler) TeamAdd(w http.ResponseWriter, r *http.Request) {
	var req teamAddRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.TeamName == "" {
		http.Error(w, "team name is empty", http.StatusBadRequest)
		return
	}

	team := model.Team{
		Name:  req.TeamName,
		Users: make([]model.User, 0, len(req.Members)),
	}

	for _, m := range req.Members {
		if m.UserID == "" || m.Username == "" {
			http.Error(w, "invalid user", http.StatusBadRequest)
			return
		}
		team.Users = append(team.Users, model.User{
			ID:       m.UserID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		})
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.teamService.CreateTeam(ctx, team); err != nil {
		if errors.Is(err, repository.ErrTeamExists) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorBody{
					Code:    "TEAM_EXISTS",
					Message: "team_name already exists",
				},
			})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 201
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := map[string]any{
		"team": map[string]any{
			"team_name": req.TeamName,
			"members":   req.Members,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// /team/get

func (h *TeamHandler) TeamGet(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, "team_name is empty", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	team, err := h.teamService.GetTeam(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
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

	members := make([]map[string]any, 0, len(team.Users))
	for _, u := range team.Users {
		members = append(members, map[string]any{
			"user_id":   u.ID,
			"username":  u.Username,
			"is_active": u.IsActive,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"team_name": team.Name,
		"members":   members,
	})

}
