package pr

import (
	"context"
	"database/sql"
	"fmt"
	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	def "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository"
	"time"
)

var _ def.PullRequestRepository = (*repository)(nil)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, req model.PullRequestPayload) (*model.PullRequest, error) {
	getCommandQuery := `
		SELECT team_name 
		FROM users 
		WHERE user_id = $1
	`

	var teamName string
	err := r.db.
		QueryRowContext(ctx, getCommandQuery, req.AuthorID).
		Scan(&teamName)

	if err != nil {
		return nil, model.ErrNotFound
	}

	getReviewiersQuery := `
		SELECT user_id FROM users
		WHERE team_name = $1 AND is_active = TRUE AND user_id <> $2
		ORDER BY random() 
		LIMIT 2
	`

	rows, err := r.db.
		QueryContext(ctx, getReviewiersQuery, teamName, req.AuthorID)

	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %v", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, id)
	}

	addNewPRQuery := `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
	`
	addReviewiers := `
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id) 
		VALUES ($1, $2)
	`
	now := time.Now().UTC()

	_, err = r.db.
		ExecContext(ctx, addNewPRQuery, req.PullRequestID, req.PullRequestName, req.AuthorID, statusToID("OPEN"), now)

	if err != nil {
		return nil, fmt.Errorf("insert pr failed: %v", err)
	}
	for _, rid := range reviewers {
		_, err := r.db.ExecContext(ctx, addReviewiers, req.PullRequestID, rid)
		if err != nil {
			return nil, fmt.Errorf("insert reviewer failed: %v", err)
		}
	}

	pr := &model.PullRequest{
		PullRequestShort: model.PullRequestShort{
			PullRequestID:   req.PullRequestID,
			PullRequestName: req.PullRequestName,
			AuthorID:        req.AuthorID,
			Status:          "OPEN",
		},
		AssignedReviewers: reviewers,
		CreatedAt:         &now,
	}
	return pr, nil
}

func (r *repository) Merge(ctx context.Context, prID string) (*model.PullRequest, error) {
	var (
		name, author        string
		statusID            int
		createdAt, mergedAt *time.Time
	)

	getPrQuery := ` 
		SELECT pull_request_name, author_id, status_id, createdAt, mergedAt 
		FROM pull_requests 
		WHERE pull_request_id = $1
	`
	err := r.db.
		QueryRowContext(ctx, getPrQuery, prID).
		Scan(&name, &author, &statusID, &createdAt, &mergedAt)

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("select pr error: %v", err)
	}

	getReviewerIdQuery := `
		SELECT reviewer_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1
	`
	rows, err := r.db.
		QueryContext(ctx, getReviewerIdQuery, prID)

	if err != nil {
		return nil, fmt.Errorf("get reviewers error: %v", err)
	}
	defer rows.Close()

	reviewers := make([]string, 0)
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, rid)
	}

	if idToStatus(statusID) == "MERGED" {
		pr := &model.PullRequest{
			PullRequestShort: model.PullRequestShort{
				PullRequestID:   prID,
				PullRequestName: name,
				AuthorID:        author,
				Status:          "MERGED",
			},
			AssignedReviewers: reviewers,
			CreatedAt:         createdAt,
			MergedAt:          mergedAt,
		}
		return pr, nil
	}

	updatePrQuery := `
        UPDATE pull_requests SET status_id=$1, mergedAt=$2 
		WHERE pull_request_id=$3
    `
	mergedNow := time.Now().UTC()
	_, err = r.db.ExecContext(ctx, updatePrQuery, statusToID("MERGED"), mergedNow, prID)
	if err != nil {
		return nil, fmt.Errorf("merge error: %v", err)
	}

	pr := &model.PullRequest{
		PullRequestShort: model.PullRequestShort{
			PullRequestID:   prID,
			PullRequestName: name,
			AuthorID:        author,
			Status:          "MERGED",
		},
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
		MergedAt:          &mergedNow,
	}
	return pr, nil
}

func (r *repository) Reassign(ctx context.Context, prID string, oldReviewerID string) (*model.PullRequest, error) {
	var (
		name, author        string
		statusID            int
		createdAt, mergedAt *time.Time
	)
	getPrQuery := `
		SELECT pull_request_name, author_id, status_id, createdAt, mergedAt
        FROM pull_requests
		WHERE pull_request_id = $1
	`

	err := r.db.
		QueryRowContext(ctx, getPrQuery, prID).
		Scan(&name, &author, &statusID, &createdAt, &mergedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select pr error: %v", err)
	}

	if idToStatus(statusID) == "MERGED" {
		return nil, model.ErrPrMerged
	}

	getReviewerIdQuery := `
		SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1
	`

	rows, err := r.db.
		QueryContext(ctx, getReviewerIdQuery, prID)

	if err != nil {
		return nil, fmt.Errorf("get reviewers error: %v", err)
	}
	defer rows.Close()

	var reviewers []string
	found := false
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, rid)
		if rid == oldReviewerID {
			found = true
		}
	}
	if !found {
		return nil, model.ErrNotFound
	}

	getTeamNameQuery := `
		SELECT team_name FROM users WHERE user_id = $1
	`
	var teamName string
	err = r.db.
		QueryRowContext(ctx, getTeamNameQuery, author).
		Scan(&teamName)

	if err != nil {
		return nil, model.ErrNotFound
	}

	getNewReviewerQuery := `
        SELECT user_id
        FROM users
        WHERE team_name = $1 AND is_active = TRUE AND user_id <> $2 AND user_id <> $3
        ORDER BY random() 
		LIMIT 1
	`
	var newReviewer string
	err = r.db.
		QueryRowContext(ctx, getNewReviewerQuery, teamName, author, oldReviewerID).
		Scan(&newReviewer)

	if err == sql.ErrNoRows {
		newReviewer = oldReviewerID
	} else if err != nil {
		return nil, fmt.Errorf("get new reviewer error: %v", err)
	}

	updateReviewerQuery := `
		UPDATE pr_reviewers
		SET reviewer_id = $1
        WHERE pull_request_id = $2 AND reviewer_id = $3
	`
	if newReviewer != oldReviewerID {
		_, err = r.db.
			ExecContext(ctx, updateReviewerQuery, newReviewer, prID, oldReviewerID)

		if err != nil {
			return nil, fmt.Errorf("update reviewer error: %v", err)
		}

		for i := range reviewers {
			if reviewers[i] == oldReviewerID {
				reviewers[i] = newReviewer
				break
			}
		}
	}

	pr := &model.PullRequest{
		PullRequestShort: model.PullRequestShort{
			PullRequestID:   prID,
			PullRequestName: name,
			AuthorID:        author,
			Status:          idToStatus(statusID),
		},
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}

	return pr, nil
}

func statusToID(status string) int {
	switch status {
	case "OPEN":
		return 1
	case "MERGED":
		return 2
	default:
		return 0
	}
}

func idToStatus(statusID int) string {
	switch statusID {
	case 1:
		return "OPEN"
	case 2:
		return "MERGED"
	default:
		return "UNKNOWN"
	}
}
