package pr

import (
	"context"
	"errors"
	"testing"

	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"time"
)

func TestCreateSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	req := model.PullRequestPayload{
		PullRequestID:   "pr-1001",
		PullRequestName: "Add search",
		AuthorID:        "u1",
	}

	mock.
		ExpectQuery("SELECT team_name FROM users WHERE user_id").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"team_name"}).AddRow("backend"))

	mock.
		ExpectQuery("SELECT user_id FROM users WHERE team_name = \\$1 AND is_active = TRUE AND user_id <>").
		WithArgs("backend", "u1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("u2").AddRow("u3"))

	mock.
		ExpectExec("INSERT INTO pull_requests").
		WithArgs("pr-1001", "Add search", "u1", statusToID("OPEN"), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec("INSERT INTO pr_reviewers").
		WithArgs("pr-1001", "u2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec("INSERT INTO pr_reviewers").
		WithArgs("pr-1001", "u3").
		WillReturnResult(sqlmock.NewResult(0, 1))

	pr, err := repo.Create(context.Background(), req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if pr == nil {
		t.Fatalf("expected PR object")
	}

	if len(pr.AssignedReviewers) != 2 || pr.AssignedReviewers[0] != "u2" || pr.AssignedReviewers[1] != "u3" {
		t.Errorf("wrong reviewers: %v", pr.AssignedReviewers)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCreateNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}
	req := model.PullRequestPayload{
		PullRequestID:   "pr-404",
		PullRequestName: "Fail",
		AuthorID:        "oops",
	}

	mock.
		ExpectQuery("SELECT team_name FROM users WHERE user_id").
		WithArgs("oops").
		WillReturnRows(sqlmock.NewRows([]string{"team_name"}))

	pr, err := repo.Create(context.Background(), req)
	if pr != nil {
		t.Errorf("must be nil PR")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMergeOpenToMerged(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}
	prID := "pr-1001"
	created := time.Now().Add(-2 * time.Hour)
	reviewers := []string{"u2", "u3"}

	mock.
		ExpectQuery("SELECT pull_request_name, author_id, status_id, createdAt, mergedAt FROM pull_requests").
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{
			"pull_request_name", "author_id", "status_id", "createdAt", "mergedAt",
		}).AddRow("Add search", "u1", statusToID("OPEN"), created, nil))

	reviewerRows := sqlmock.NewRows([]string{"reviewer_id"}).AddRow(reviewers[0]).AddRow(reviewers[1])
	mock.
		ExpectQuery("SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id").
		WithArgs(prID).
		WillReturnRows(reviewerRows)

	mock.
		ExpectExec("UPDATE pull_requests SET").
		WithArgs(statusToID("MERGED"), sqlmock.AnyArg(), prID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	pr, err := repo.Merge(context.Background(), prID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if pr == nil {
		t.Fatalf("expected non-nil PR")
	}
	if pr.Status != "MERGED" {
		t.Errorf("expected status MERGED, got %v", pr.Status)
	}
	if len(pr.AssignedReviewers) != 2 || pr.AssignedReviewers[0] != "u2" {
		t.Errorf("wrong reviewers: %v", pr.AssignedReviewers)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMergeNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}
	prID := "pr-404"

	mock.
		ExpectQuery("SELECT pull_request_name, author_id, status_id, createdAt, mergedAt FROM pull_requests").
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{
			"pull_request_name", "author_id", "status_id", "createdAt", "mergedAt",
		}))

	pr, err := repo.Merge(context.Background(), prID)
	if pr != nil {
		t.Errorf("expected nil, got %+v", pr)
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMergeAlreadyMerged(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}
	prID := "pr-1001"
	created := time.Now().Add(-2 * time.Hour)
	merged := time.Now().Add(-1 * time.Hour)
	reviewers := []string{"u2", "u3"}

	mock.
		ExpectQuery("SELECT pull_request_name, author_id, status_id, createdAt, mergedAt FROM pull_requests").
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{
			"pull_request_name", "author_id", "status_id", "createdAt", "mergedAt",
		}).AddRow("Add search", "u1", statusToID("MERGED"), created, merged))

	reviewerRows := sqlmock.NewRows([]string{"reviewer_id"}).AddRow(reviewers[0]).AddRow(reviewers[1])
	mock.
		ExpectQuery("SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id").
		WithArgs(prID).
		WillReturnRows(reviewerRows)

	pr, err := repo.Merge(context.Background(), prID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if pr == nil {
		t.Fatalf("expected non-nil PR, got nil")
	}
	if pr.Status != "MERGED" {
		t.Errorf("wrong status %v", pr.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestReassignSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	prID := "pr-1001"
	oldReviewer := "u2"
	newReviewer := "u5"
	author := "u1"

	created := time.Now().Add(-1 * time.Hour)

	mock.
		ExpectQuery("SELECT pull_request_name, author_id, status_id, createdAt, mergedAt FROM pull_requests").
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows(
			[]string{"pull_request_name", "author_id", "status_id", "createdAt", "mergedAt"},
		).AddRow("Add search", author, statusToID("OPEN"), created, nil))

	mock.
		ExpectQuery("SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id").
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"reviewer_id"}).AddRow("u3").AddRow(oldReviewer))

	mock.
		ExpectQuery("SELECT team_name FROM users WHERE user_id").
		WithArgs(author).
		WillReturnRows(sqlmock.NewRows([]string{"team_name"}).AddRow("backend"))

	mock.
		ExpectQuery("SELECT user_id FROM users WHERE team_name = \\$1 AND is_active = TRUE AND user_id <>").
		WithArgs("backend", author, oldReviewer).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(newReviewer))

	mock.
		ExpectExec("UPDATE pr_reviewers SET reviewer_id = \\$1 WHERE pull_request_id").
		WithArgs(newReviewer, prID, oldReviewer).
		WillReturnResult(sqlmock.NewResult(0, 1))

	pr, _, err := repo.Reassign(context.Background(), prID, oldReviewer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if pr == nil {
		t.Fatalf("expected non-nil PR")
	}
	if len(pr.AssignedReviewers) != 2 || pr.AssignedReviewers[0] != "u3" || pr.AssignedReviewers[1] != newReviewer {
		t.Errorf("wrong reviewers after reassign: %+v", pr.AssignedReviewers)
	}
	if pr.Status != "OPEN" {
		t.Errorf("status must be OPEN")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
