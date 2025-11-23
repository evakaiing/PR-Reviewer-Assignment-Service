package team

import (
	"context"
	"errors"
	"testing"

	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestAddSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	team := &model.Team{
		TeamName: "backend",
		Members: []*model.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	mock.
		ExpectBegin()

	mock.
		ExpectQuery("INSERT INTO teams").
		WithArgs("backend").
		WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

	mock.
        ExpectExec("INSERT INTO users").
        WithArgs("u1", "Alice", "backend", true).
        WillReturnResult(sqlmock.NewResult(0, 1))


	mock.
        ExpectExec("INSERT INTO users").
        WithArgs("u2", "Bob", "backend", true).
        WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectCommit()

	err = repo.Add(context.Background(), team)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAddTeamExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	team := &model.Team{
		TeamName: "backend",
		Members:  []*model.TeamMember{{UserID: "u1"}},
	}

	mock.
		ExpectBegin()

	mock.
		ExpectQuery("INSERT INTO teams").
		WithArgs("backend").
		WillReturnRows(sqlmock.NewRows([]string{"team_id"}))

	mock.
		ExpectRollback()

	err = repo.Add(context.Background(), team)

	if !errors.Is(err, model.ErrTeamExists) {
		t.Errorf("expected ErrTeamExists, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TesAddBeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	team := &model.Team{
		TeamName: "backend",
		Members:  []*model.TeamMember{{UserID: "u1"}},
	}

	mock.
		ExpectBegin().
		WillReturnError(errors.New("connection error"))

	err = repo.Add(context.Background(), team)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAddCreateTeamError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	team := &model.Team{
		TeamName: "backend",
		Members:  []*model.TeamMember{{UserID: "u1"}},
	}

	mock.
		ExpectBegin()

	mock.
		ExpectQuery("INSERT INTO teams").
		WithArgs("backend").
		WillReturnError(errors.New("database error"))

	mock.
		ExpectRollback()

	err = repo.Add(context.Background(), team)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	teamName := "backend"

	rows := sqlmock.NewRows([]string{"user_id", "username", "is_active"}).
		AddRow("u1", "Alice", true).
		AddRow("u2", "Bob", true)

	mock.
		ExpectQuery("SELECT user_id, username, is_active FROM users WHERE team_name").
		WithArgs(teamName).
		WillReturnRows(rows)

	team, err := repo.Get(context.Background(), teamName)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if team == nil {
		t.Fatalf("expected non-nil team")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetTeamNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	teamName := "unknownteam"

	emptyRows := sqlmock.NewRows([]string{"user_id", "username", "is_active"})

	mock.
		ExpectQuery("SELECT user_id, username, is_active FROM users WHERE team_name").
		WithArgs(teamName).
		WillReturnRows(emptyRows)

	team, err := repo.Get(context.Background(), teamName)
	if team != nil {
		t.Errorf("expected nil team, got %+v", team)
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("expected ErrTeamNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := &repository{db: db}

	teamName := "backend"

	mock.
		ExpectQuery("SELECT user_id, username, is_active FROM users WHERE team_name").
		WithArgs(teamName).
		WillReturnError(errors.New("connection lost"))

	team, err := repo.Get(context.Background(), teamName)
	if team != nil {
		t.Errorf("expected nil team, got %+v", team)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
