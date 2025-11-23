package service

import (
	"context"

	"github.com/Olzerq/avito-pr-reviewer/internal/repository"
)

type StatsService struct {
	statsRepo *repository.StatsRepository
}

func NewStatsService(statsRepo *repository.StatsRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo}
}

type AssignmentsStats struct {
	ByUser []repository.UserAssignmentsStat
	ByPR   []repository.PullRequestAssignmentsStat
}

func (s *StatsService) GetAssignmentsStats(ctx context.Context) (AssignmentsStats, error) {
	users, err := s.statsRepo.GetAssignmentsByUser(ctx)
	if err != nil {
		return AssignmentsStats{}, err
	}

	prs, err := s.statsRepo.GetAssignmentsByPR(ctx)
	if err != nil {
		return AssignmentsStats{}, err
	}

	return AssignmentsStats{
		ByUser: users,
		ByPR:   prs,
	}, nil
}
