package repository

import (
	"context"
	"errors"
	"github.com/Olzerq/avito-pr-reviewer/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTeamExists = errors.New("team already exists")
var ErrTeamNotFound = errors.New("team not found")

type TeamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{db: db}
}

// CreateTeam - создает команду
func (r *TeamRepository) CreateTeam(ctx context.Context, team model.Team) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Пытаемся создать команду
	_, err = tx.Exec(ctx,
		`INSERT INTO teams (team_name)
         VALUES ($1)`,
		team.Name,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 — код ошибки на уникальность
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrTeamExists
		}
		return err
	}

	for _, u := range team.Users {
		_, err = tx.Exec(ctx,
			`INSERT INTO users (user_id, username, team_name, is_active)
             VALUES ($1, $2, $3, $4)
             ON CONFLICT (user_id) DO UPDATE
             SET username = EXCLUDED.username,
                 team_name = EXCLUDED.team_name,
                 is_active = EXCLUDED.is_active`,
			u.ID, u.Username, team.Name, u.IsActive,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetTeam возвращает команду с ее пользователями
func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (model.Team, error) {
	var exists bool
	// проверяем существование команды
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
             SELECT 1 FROM teams WHERE team_name = $1
         )`,
		teamName,
	).Scan(&exists)
	if err != nil {
		return model.Team{}, err
	}
	if !exists {
		return model.Team{}, ErrTeamNotFound
	}

	rows, err := r.db.Query(ctx,
		`SELECT user_id, username, team_name, is_active
         FROM users
         WHERE team_name = $1`,
		teamName,
	)
	if err != nil {
		return model.Team{}, err
	}
	defer rows.Close()

	team := model.Team{
		Name:  teamName,
		Users: make([]model.User, 0),
	}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return model.Team{}, err
		}
		team.Users = append(team.Users, u)
	}

	if err := rows.Err(); err != nil {
		return model.Team{}, err
	}
	return team, nil
}
