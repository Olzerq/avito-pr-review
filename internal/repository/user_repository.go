package repository

import (
	"context"
	"errors"
	"github.com/Olzerq/avito-pr-reviewer/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// SetIsActive устанавливает флаг активности пользователя
func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	cmdTag, err := r.db.Exec(ctx,
		`UPDATE users
         SET is_active = $1
         WHERE user_id = $2`,
		isActive, userID,
	)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, userID string) (model.User, error) {
	var u model.User

	err := r.db.QueryRow(ctx,
		`SELECT user_id, username, team_name, is_active
         FROM users
         WHERE user_id = $1`,
		userID,
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}

	return u, nil
}

// GetActiveByTeamExcept возвращает активных пользователей команды исключая авторов и уже назначенных
func (r *UserRepository) GetActiveByTeamExcept(ctx context.Context, teamName string, excludeIDs []string) ([]model.User, error) {
	// Если исключать некого
	if len(excludeIDs) == 0 {
		rows, err := r.db.Query(ctx,
			`SELECT user_id, username, team_name, is_active
             FROM users
             WHERE team_name = $1
               AND is_active = TRUE`,
			teamName,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var res []model.User
		for rows.Next() {
			var u model.User
			if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
				return nil, err
			}
			res = append(res, u)
		}
		return res, rows.Err()
	}

	rows, err := r.db.Query(ctx,
		`SELECT user_id, username, team_name, is_active
         FROM users
         WHERE team_name = $1
           AND is_active = TRUE
           AND user_id <> ALL($2)`,
		teamName, excludeIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, rows.Err()
}
