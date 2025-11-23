package service

import (
	"context"
	"github.com/Olzerq/avito-pr-reviewer/internal/model"
	"github.com/Olzerq/avito-pr-reviewer/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	return s.userRepo.SetIsActive(ctx, userID, isActive)
}

// GetByID

func (s *UserService) GetByID(ctx context.Context, userID string) (model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
