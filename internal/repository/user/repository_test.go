package user

import (
	"context"
	"errors"
	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
)

func TestSetIsActiveSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	mock.
		ExpectExec("UPDATE users SET is_active").
		WithArgs(false, "user123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id = \\$1").
		WithArgs("user123").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
			AddRow("user123", "Bob", "backend", false))

	_, err = repo.SetIsActive(context.Background(), "user123", false)

	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSetIsActiveUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	mock.
		ExpectExec("UPDATE users SET is_active").
		WithArgs(true, "none").
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err = repo.SetIsActive(context.Background(), "none", true)

	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSetIsActiveDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	mock.
		ExpectExec("UPDATE users SET is_active").
		WithArgs(true, "user123").
		WillReturnError(errors.New("connection lost"))

	_, err = repo.SetIsActive(context.Background(), "user123", true)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetReviewSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	prRows := sqlmock.NewRows([]string{
		"pull_request_id", "pull_request_name", "author_id",
		"status_name",
	}).
		AddRow("pr-1001", "Add search", "u1", "OPEN").
		AddRow("pr-1002", "Add readme", "u2", "MERGED")

	mock.
		ExpectQuery("SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, ps.status_name FROM pull_requests pr").
		WithArgs("reviewer1").
		WillReturnRows(prRows)

	prs, err := repo.GetReview(context.Background(), "reviewer1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prs) != 2 {
		t.Errorf("expected 2 prs, got %d", len(prs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
