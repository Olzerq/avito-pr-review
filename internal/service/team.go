package service

import (
	"context"
	"github.com/Olzerq/avito-pr-reviewer/internal/model"
	"github.com/Olzerq/avito-pr-reviewer/internal/repository"
)

type TeamService struct {
	teamRepo *repository.TeamRepository
}

func NewTeamService(teamRepo *repository.TeamRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo}
}

// CreateTeam создает команду
func (s *TeamService) CreateTeam(ctx context.Context, team model.Team) error {
	return s.teamRepo.CreateTeam(ctx, team)
}

// GetTeam возвращает команду
func (s *TeamService) GetTeam(ctx context.Context, teamName string) (model.Team, error) {
	return s.teamRepo.GetTeam(ctx, teamName)
}
