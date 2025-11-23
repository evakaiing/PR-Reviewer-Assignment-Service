package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
)

var ErrorMap = map[error]struct {
	Code    string
	Message string
}{
	model.ErrTeamExists:   {"TEAM_EXISTS", "team_name already exists"},
	model.ErrPrExists:     {"PR_EXISTS", "PR id already exists"},
	model.ErrPrMerged:     {"PR_MERGED", "cannot reassign on merged PR"},
	model.ErrNotAssigned:  {"NOT_ASSIGNED", "reviewer is not assigned to this PR"},
	model.ErrNoCandidate:  {"NO_CANDIDATE", "no active replacement candidate in team"},
	model.ErrNotFound:     {"NOT_FOUND", "resource not found"},
	model.ErrInvalidInput: {"INVALID_REQUEST", "invalid request body"},
	model.ErrMissingParam: {"INVALID_REQUEST", "missing required parameter"},
}

type BaseHandler struct {
	Logger slog.Logger
}

func (h *BaseHandler) WriteErrorFromMap(w http.ResponseWriter, err error, status int, info ...any) {
	resp := model.ErrorResponse{}
	if data, ok := ErrorMap[err]; ok {
		resp.Error.Code = data.Code
		resp.Error.Message = data.Message
	} else {
		resp.Error.Code = "UNKNOWN_ERROR"
		resp.Error.Message = err.Error()
	}
	h.Logger.Error("API error",
		slog.Int("http_status", status),
		slog.String("error_code", resp.Error.Code),
		slog.String("error_message", resp.Error.Message),
		slog.Any("info", info),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *BaseHandler) WriteJSON(w http.ResponseWriter, v interface{}, status int, info ...any) {
	h.Logger.Info("API success response",
		slog.Int("http_status", status),
		slog.Any("body", v),
		slog.Any("info", info),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
