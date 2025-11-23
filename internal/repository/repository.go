package repository

import (
	"context"
	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
)

//go:generate mockgen -source=repository.go -destination=../mocks/mock_repository.go -package=mocks

type TeamRepository interface {
	Add(ctx context.Context, team *model.Team) error
	Get(ctx context.Context, teamName string) (*model.Team, error)
}

type UserRepository interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error)
	GetReview(ctx context.Context, userID string) ([]*model.PullRequestShort, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, req model.PullRequestPayload) (*model.PullRequest, error)
	Merge(ctx context.Context, prID string) (*model.PullRequest, error)
	Reassign(ctx context.Context, prID string, oldReviewerID string) (*model.PullRequest, string, error)
}
