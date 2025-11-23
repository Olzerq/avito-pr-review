package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserAssignmentsStat struct {
	UserID        string
	AssignedCount int
}

type PullRequestAssignmentsStat struct {
	PullRequestID  string
	ReviewersCount int
}

type StatsRepository struct {
	db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{db: db}
}

// GetAssignmentsByUser Возвращает количество назначений по юзерам
func (r *StatsRepository) GetAssignmentsByUser(ctx context.Context) ([]UserAssignmentsStat, error) {
	rows, err := r.db.Query(ctx,
		`SELECT user_id, COUNT(*) AS assigned_count
         FROM pull_request_reviewers
         GROUP BY user_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]UserAssignmentsStat, 0)
	for rows.Next() {
		var s UserAssignmentsStat
		if err := rows.Scan(&s.UserID, &s.AssignedCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetAssignmentsByPR Возвращает количество ревьюверов по PR
func (r *StatsRepository) GetAssignmentsByPR(ctx context.Context) ([]PullRequestAssignmentsStat, error) {
	rows, err := r.db.Query(ctx,
		`SELECT pull_request_id, COUNT(*) AS reviewers_count
         FROM pull_request_reviewers
         GROUP BY pull_request_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]PullRequestAssignmentsStat, 0)
	for rows.Next() {
		var s PullRequestAssignmentsStat
		if err := rows.Scan(&s.PullRequestID, &s.ReviewersCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
