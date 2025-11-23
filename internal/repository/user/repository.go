package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	def "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository"
)

var _ def.UserRepository = (*repository)(nil)

var (
	ErrUserNotFound = errors.New("user not found")
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r *repository) SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	query := `
        UPDATE users
        SET is_active=$1, updated_at = CURRENT_TIMESTAMP
        WHERE user_id=$2
    `
	result, err := r.db.ExecContext(ctx, query, isActive, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user active status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	var user model.User
	getQuery := `
        SELECT user_id, username, team_name, is_active
        FROM users
        WHERE user_id = $1
    `
	err = r.db.QueryRowContext(ctx, getQuery, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %v", err)
	}

	return &user, nil
}

func (r *repository) GetReview(ctx context.Context, reviewerID string) ([]*model.PullRequestShort, error) {
	query := `
        SELECT 
            pr.pull_request_id,
            pr.pull_request_name,
            pr.author_id,
            ps.status_name
        FROM pull_requests pr
        INNER JOIN pull_request_statuses ps 
            ON pr.status_id = ps.status_id
        INNER JOIN pull_request_reviewers prr 
            ON pr.pull_request_id = prr.pull_request_id
        WHERE prr.reviewer_user_id = $1
        ORDER BY pr.created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to select pull requests with current reviewer id: %v", err)
	}
	defer rows.Close()

	pullRequests := make([]*model.PullRequestShort, 0)

	for rows.Next() {
		pr := &model.PullRequestShort{}
		err = rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %v", err)
		}
		pullRequests = append(pullRequests, pr)
	}

	return pullRequests, nil
}
