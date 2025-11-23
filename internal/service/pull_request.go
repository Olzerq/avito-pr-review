package service

import (
	"context"
	"errors"
	"github.com/Olzerq/avito-pr-reviewer/internal/model"
	"github.com/Olzerq/avito-pr-reviewer/internal/repository"
	"math/rand"
	"time"
)

var (
	ErrPRMerged    = errors.New("pull request already merged")
	ErrNoCandidate = errors.New("no candidate reviewer found")
)

type PullRequestService struct {
	prRepo   *repository.PullRequestRepository
	userRepo *repository.UserRepository
}

func NewPullRequestService(prRepo *repository.PullRequestRepository, userRepo *repository.UserRepository) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

// Create создаёт PR и выбирает случайных ревьюеров
func (s *PullRequestService) Create(ctx context.Context, id, name, authorID string) (model.PullRequest, error) {
	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return model.PullRequest{}, err
		}
		return model.PullRequest{}, err
	}

	candidates, err := s.userRepo.GetActiveByTeamExcept(ctx, author.TeamName, []string{author.ID})
	if err != nil {
		return model.PullRequest{}, err
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	n := len(candidates)
	if n > 2 {
		n = 2
	}
	reviewers := candidates[:n]

	pr := model.PullRequest{
		ID:        id,
		Name:      name,
		AuthorID:  author.ID,
		Status:    "OPEN",
		MergedAt:  nil,
		Reviewers: reviewers,
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return model.PullRequest{}, err
	}

	return pr, nil
}

// Reassign переназначает ревьюера
func (s *PullRequestService) Reassign(
	ctx context.Context,
	prID string,
	oldReviewerID string,
) (model.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return model.PullRequest{}, "", err // ErrPRNotFound пойдёт наверх
	}

	if pr.Status == "MERGED" {
		return model.PullRequest{}, "", ErrPRMerged
	}

	isAssigned := false
	for _, r := range pr.Reviewers {
		if r.ID == oldReviewerID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return model.PullRequest{}, "", repository.ErrReviewerNotAssigned
	}

	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return model.PullRequest{}, "", err // ErrUserNotFound → 404
	}

	exclude := make([]string, 0, len(pr.Reviewers)+1)
	exclude = append(exclude, pr.AuthorID)
	for _, r := range pr.Reviewers {
		exclude = append(exclude, r.ID)
	}

	candidates, err := s.userRepo.GetActiveByTeamExcept(ctx, author.TeamName, exclude)
	if err != nil {
		return model.PullRequest{}, "", err
	}
	if len(candidates) == 0 {
		return model.PullRequest{}, "", ErrNoCandidate
	}

	rand.Seed(time.Now().UnixNano())
	newReviewer := candidates[rand.Intn(len(candidates))]

	if err := s.prRepo.ReassignReviewer(ctx, pr.ID, oldReviewerID, newReviewer.ID); err != nil {
		return model.PullRequest{}, "", err
	}

	updatedPR, err := s.prRepo.GetByID(ctx, pr.ID)
	if err != nil {
		return model.PullRequest{}, "", err
	}

	return updatedPR, newReviewer.ID, nil
}

// Merge обновляет флаг Merged
func (s *PullRequestService) Merge(ctx context.Context, prID string) (model.PullRequest, error) {
	// проверяем, что PR существует
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return model.PullRequest{}, err
	}
	//если уже merged то возвращаем как есть - идемпотентность
	if pr.Status != "MERGED" {
		if err := s.prRepo.MarkMerged(ctx, prID, time.Now().UTC()); err != nil {
			return model.PullRequest{}, err
		}
	}

	updated, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return model.PullRequest{}, err
	}
	return updated, nil
}

// GetUserReviews получает все пры где юзер ревьювер
func (s *PullRequestService) GetUserReviews(ctx context.Context, userID string) ([]model.PullRequestShort, error) {
	// проверяем что юзер существует
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, repository.ErrUserNotFound
	}

	return s.prRepo.GetByReviewer(ctx, userID)
}
