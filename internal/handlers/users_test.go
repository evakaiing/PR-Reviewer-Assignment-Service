package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/mocks"
	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	"go.uber.org/mock/gomock"
)

func TestSetIsActiveSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	handler := &UserHandler{
		BaseHandler: BaseHandler{
			Logger: *slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		UserRepo: mockRepo,
	}

	expectedUser := &model.User{
		TeamMember: model.TeamMember{
			UserID:   "u1",
			Username: "Alice",
			IsActive: false,
		},
		TeamName: "backend",
	}

	mockRepo.
		EXPECT().
		SetIsActive(gomock.Any(), "u1", false).
		Return(expectedUser, nil)

	reqBody := map[string]any{
		"user_id":   "u1",
		"is_active": false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.SetIsActive(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
		return
	}

	var result map[string]any
	json.Unmarshal(respBody, &result)

	user, ok := result["user"].(map[string]any)
	if !ok {
		t.Errorf("expected user object in response")
		return
	}

	if user["user_id"] != "u1" {
		t.Errorf("expected user_id u1, got %v", user["user_id"])
		return
	}
}

func TestSetIsActiveUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	handler := &UserHandler{
		BaseHandler: BaseHandler{
			Logger: *slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		UserRepo: mockRepo,
	}

	mockRepo.
		EXPECT().
		SetIsActive(gomock.Any(), "none", true).
		Return(nil, model.ErrNotFound)

	reqBody := map[string]any{
		"user_id":   "none",
		"is_active": true,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.SetIsActive(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
		return
	}
}

func TestGetReviewSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	handler := &UserHandler{
		BaseHandler: BaseHandler{
			Logger: *slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		UserRepo: mockRepo,
	}

	expectedPRs := []*model.PullRequestShort{
		{
			PullRequestID:   "pr-1001",
			PullRequestName: "Add search",
			AuthorID:        "u1",
			Status:          "OPEN",
		},
		{
			PullRequestID:   "pr-1002",
			PullRequestName: "Fix bug",
			AuthorID:        "u2",
			Status:          "MERGED",
		},
	}

	mockRepo.
		EXPECT().
		GetReview(gomock.Any(), "u1").
		Return(expectedPRs, nil)

	req := httptest.NewRequest("GET", "/users/getReview?user_id=u1", nil)
	w := httptest.NewRecorder()

	handler.GetReview(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
		return
	}

	var result map[string]any
	json.Unmarshal(respBody, &result)

	if result["user_id"] != "u1" {
		t.Errorf("expected user_id u1, got %v", result["user_id"])
		return
	}
}

func TestGetReviewMissingUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	handler := &UserHandler{
		BaseHandler: BaseHandler{
			Logger: *slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		UserRepo: mockRepo,
	}

	req := httptest.NewRequest("GET", "/users/getReview", nil)
	w := httptest.NewRecorder()

	handler.GetReview(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
		return
	}
}

func TestGetReviewUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	handler := &UserHandler{
		BaseHandler: BaseHandler{
			Logger: *slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		UserRepo: mockRepo,
	}

	mockRepo.
		EXPECT().
		GetReview(gomock.Any(), "none").
		Return(nil, model.ErrNotFound)

	req := httptest.NewRequest("GET", "/users/getReview?user_id=none", nil)
	w := httptest.NewRecorder()

	handler.GetReview(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
		return
	}
}
