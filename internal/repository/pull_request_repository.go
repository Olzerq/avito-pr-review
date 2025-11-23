package repository

import (
	"context"
	"errors"
	"github.com/Olzerq/avito-pr-reviewer/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var (
	ErrPRExists            = errors.New("pull request already exists")
	ErrPRNotFound          = errors.New("pull request not found")
	ErrReviewerNotAssigned = errors.New("reviewer not assigned to this pull request")
)

type PullRequestRepository struct {
	db *pgxpool.Pool
}

func NewPullRequestRepository(db *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) Create(ctx context.Context, pr model.PullRequest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()

	_, err = tx.Exec(ctx,
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, merged_at)
		 VALUES ($1, $2, $3, $4, NULL)`,
		pr.ID, pr.Name, pr.AuthorID, pr.Status,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrPRExists
		}
		return err
	}

	for _, u := range pr.Reviewers {
		_, err = tx.Exec(ctx,
			`INSERT INTO pull_request_reviewers (pull_request_id, user_id)
             VALUES ($1, $2)`,
			pr.ID, u.ID,
		)
		if err != nil {
			return err
		}
	}

	_ = now

	return tx.Commit(ctx)
}

// GetByID возвращает PR и его ревьюеров
func (r *PullRequestRepository) GetByID(ctx context.Context, prID string) (model.PullRequest, error) {
	var pr model.PullRequest

	err := r.db.QueryRow(ctx,
		`SELECT pull_request_id, pull_request_name, author_id, status, merged_at
         FROM pull_requests
         WHERE pull_request_id = $1`,
		prID,
	).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.PullRequest{}, ErrPRNotFound
		}
		return model.PullRequest{}, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT u.user_id, u.username, u.team_name, u.is_active
         FROM pull_request_reviewers prr
         JOIN users u ON prr.user_id = u.user_id
         WHERE prr.pull_request_id = $1`,
		prID,
	)
	if err != nil {
		return model.PullRequest{}, err
	}
	defer rows.Close()

	pr.Reviewers = make([]model.User, 0)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return model.PullRequest{}, err
		}
		pr.Reviewers = append(pr.Reviewers, u)
	}

	return pr, rows.Err()
}

// ReassignReviewer Переназначает ревьюера
func (r *PullRequestRepository) ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// удаляем старого ревьюера
	cmdTag, err := tx.Exec(ctx,
		`DELETE FROM pull_request_reviewers
         WHERE pull_request_id = $1 AND user_id = $2`,
		prID, oldUserID,
	)
	if err != nil {
		return err
	}

	// если он не был назначен
	if cmdTag.RowsAffected() == 0 {
		return ErrReviewerNotAssigned
	}

	// добавляем нового
	_, err = tx.Exec(ctx,
		`INSERT INTO pull_request_reviewers (pull_request_id, user_id)
         VALUES ($1, $2)`,
		prID, newUserID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// MarkMerged обновляет флаг Merged
func (r *PullRequestRepository) MarkMerged(ctx context.Context, prID string, mergedAt time.Time) error {
	cmdTag, err := r.db.Exec(ctx,
		`UPDATE pull_requests
         SET status = 'MERGED',
             merged_at = COALESCE(merged_at, $2)
         WHERE pull_request_id = $1`,
		prID, mergedAt,
	)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrPRNotFound
	}

	return nil
}

// GetByReviewer получает PR'ы, где пользователь является ревьювером
func (r *PullRequestRepository) GetByReviewer(ctx context.Context, userID string) ([]model.PullRequestShort, error) {
	rows, err := r.db.Query(ctx,
		`SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
         FROM pull_requests pr
         JOIN pull_request_reviewers prr
           ON pr.pull_request_id = prr.pull_request_id
         WHERE prr.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.PullRequestShort
	for rows.Next() {
		var pr model.PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		result = append(result, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
