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

func TestAddSuccess(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockTeamRepository(ctrl)
    handler := &TeamHandler{
        BaseHandler: BaseHandler{
            Logger: *slog.New(slog.NewTextHandler(io.Discard, nil)),
        },
        TeamRepo: mockRepo,
    }

    reqBody := map[string]any{
        "team_name": "payments",
        "members": []map[string]any{
            {"user_id": "u1", "username": "Alice", "is_active": true},
            {"user_id": "u2", "username": "Bob", "is_active": true},
        },
    }

    body, _ := json.Marshal(reqBody)

    mockRepo.
		EXPECT().
        Add(gomock.Any(), gomock.Any()).
        Return(nil)

    expectedTeam := &model.Team{
        TeamName: "payments",
        Members: []*model.TeamMember{
            {UserID: "u1", Username: "Alice", IsActive: true},
            {UserID: "u2", Username: "Bob", IsActive: true},
        },
    }
    mockRepo.
		EXPECT().
        Get(gomock.Any(), "payments").
        Return(expectedTeam, nil)

    req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
    w := httptest.NewRecorder()

    handler.Add(w, req)

    resp := w.Result()
    respBody, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusCreated {
        t.Errorf("expected status 201, got %d", resp.StatusCode)
        return
    }

    var result map[string]any
    json.Unmarshal(respBody, &result)

    team, ok := result["team"].(map[string]any)
    if !ok {
        t.Errorf("expected team object in response")
        return
    }

    if team["team_name"] != "payments" {
        t.Errorf("expected team_name payments, got %v", team["team_name"])
        return
    }
}

func TestAddTeamExists(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockTeamRepository(ctrl)
    handler := &TeamHandler{
        BaseHandler: BaseHandler{
            Logger: *slog.New(slog.NewTextHandler(io.Discard, nil)),
        },
        TeamRepo: mockRepo,
    }

    reqBody := map[string]any{
        "team_name": "payments",
        "members": []map[string]any{
            {"user_id": "u1", "username": "Alice", "is_active": true},
        },
    }
    body, _ := json.Marshal(reqBody)

    mockRepo.
		EXPECT().
        Add(gomock.Any(), gomock.Any()).
        Return(model.ErrTeamExists)

    req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
    w := httptest.NewRecorder()

    handler.Add(w, req)

    resp := w.Result()

    if resp.StatusCode != http.StatusBadRequest {
        t.Errorf("expected status 400, got %d", resp.StatusCode)
        return
    }
}
func TestGetTeamNotFound(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockTeamRepository(ctrl)
    handler := &TeamHandler{
        BaseHandler: BaseHandler{
            Logger: *slog.New(slog.NewTextHandler(io.Discard, nil)),
        },
        TeamRepo: mockRepo,
    }

    mockRepo.EXPECT().
        Get(gomock.Any(), "none").
        Return(nil, model.ErrNotFound)

    req := httptest.NewRequest("GET", "/team/get?team_name=none", nil)
    w := httptest.NewRecorder()

    handler.Get(w, req)

    resp := w.Result()

    if resp.StatusCode != http.StatusNotFound {
        t.Errorf("expected status 404, got %d", resp.StatusCode)
        return
    }
}