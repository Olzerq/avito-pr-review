package model

import "time"

// Участник команды
type User struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// Команда
type Team struct {
	Name  string `json:"team_name"`
	Users []User `json:"users"`
}

type PullRequest struct {
	ID       string     `json:"pull_request_id"`
	Name     string     `json:"pull_request_name"`
	AuthorID string     `json:"author_id"`
	Status   string     `json:"status"` // "OPEN" или "MERGED"
	MergedAt *time.Time `json:"merged_at,omitempty"`

	Reviewers []User `json:"reviewers"`
}

type PullRequestShort struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}
