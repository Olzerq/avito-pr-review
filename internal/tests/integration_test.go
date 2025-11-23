package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Olzerq/avito-pr-reviewer/internal/config"
	httpapi "github.com/Olzerq/avito-pr-reviewer/internal/http"
	"github.com/Olzerq/avito-pr-reviewer/internal/repository"
	"github.com/Olzerq/avito-pr-reviewer/internal/service"
	"github.com/go-chi/chi/v5"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	cfg := config.LoadConfig()

	ctx := context.Background()
	db := repository.NewDB(ctx, cfg.DBConnStr())

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPullRequestRepository(db)

	// Services
	teamService := service.NewTeamService(teamRepo)
	// userService := service.NewUserService(userRepo)
	prService := service.NewPullRequestService(prRepo, userRepo)

	// Handlers
	teamHandler := httpapi.NewTeamHandler(teamService)
	// userHandler := httpapi.NewUserHandler(userService)
	prHandler := httpapi.NewPullRequestHandler(prService)

	r := chi.NewRouter()

	r.Post("/team/add", teamHandler.TeamAdd)
	r.Post("/pullRequest/create", prHandler.Create)
	r.Get("/users/getReview", prHandler.GetUserReviews)

	return httptest.NewServer(r)
}

func TestFullFlow(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	teamName := "backend_" + t.Name()

	teamBody := map[string]any{
		"team_name": teamName,
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}

	bodyBytes, err := json.Marshal(teamBody)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(server.URL+"/team/add", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 got %d", resp.StatusCode)
	}

	prReq := `{
		"pull_request_id": "pr-1",
		"pull_request_name": "Add search",
		"author_id": "u1"
	}`

	resp, err = http.Post(server.URL+"/pullRequest/create", "application/json", bytes.NewBuffer([]byte(prReq)))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 got %d", resp.StatusCode)
	}

	var prResp map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&prResp)
	resp.Body.Close()

	pr := prResp["pr"].(map[string]any)
	assigned := pr["assigned_reviewers"].([]any)
	if len(assigned) == 0 {
		t.Fatalf("expected assigned reviewers, got 0")
	}

	resp, err = http.Get(server.URL + "/users/getReview?user_id=u2")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}

	var reviewResp map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&reviewResp)

	reviews := reviewResp["pull_requests"].([]any)
	if len(reviews) == 0 {
		t.Fatalf("expected non-empty review list")
	}
}
