package handlers

import (
    "encoding/json"
    "errors"
    "log/slog"
    "net/http"

    "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
    "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository"
)

type PullRequestHandler struct {
    BaseHandler
    PRRepo repository.PullRequestRepository
}

func NewPullRequestHandler(logger slog.Logger, prRepo repository.PullRequestRepository) *PullRequestHandler {
    return &PullRequestHandler{
        BaseHandler: BaseHandler{Logger: logger},
        PRRepo:      prRepo,
    }
}

func (h *PullRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req model.PullRequestPayload
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.WriteErrorFromMap(w, model.ErrInvalidInput, http.StatusBadRequest,
            slog.String("path", r.URL.Path))
        return
    }

    if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
        h.WriteErrorFromMap(w, model.ErrMissingParam, http.StatusBadRequest,
            slog.String("fields", "pull_request_id, pull_request_name, or author_id"))
        return
    }

    pr, err := h.PRRepo.Create(r.Context(), req)
    if err != nil {
        status := http.StatusInternalServerError
        if errors.Is(err, model.ErrNotFound) {
            status = http.StatusNotFound 
        } else if errors.Is(err, model.ErrPrExists) {
            status = http.StatusConflict
        }
        h.WriteErrorFromMap(w, err, status,
            slog.String("pull_request_id", req.PullRequestID))
        return
    }

    h.WriteJSON(w, map[string]any{"pr": pr}, http.StatusCreated,
        slog.String("pull_request_id", req.PullRequestID))
}

func (h *PullRequestHandler) Merge(w http.ResponseWriter, r *http.Request) {
    type reqBody struct {
        PullRequestID string `json:"pull_request_id"`
    }
    var req reqBody
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.WriteErrorFromMap(w, model.ErrInvalidInput, http.StatusBadRequest,
            slog.String("path", r.URL.Path))
        return
    }

    if req.PullRequestID == "" {
        h.WriteErrorFromMap(w, model.ErrMissingParam, http.StatusBadRequest,
            slog.String("field", "pull_request_id"))
        return
    }

    pr, err := h.PRRepo.Merge(r.Context(), req.PullRequestID)
    if err != nil {
        status := http.StatusInternalServerError
        if errors.Is(err, model.ErrNotFound) {
            status = http.StatusNotFound
        }
        h.WriteErrorFromMap(w, err, status,
            slog.String("pull_request_id", req.PullRequestID))
        return
    }

    h.WriteJSON(w, map[string]any{"pr": pr}, http.StatusOK,
        slog.String("pull_request_id", req.PullRequestID))
}

func (h *PullRequestHandler) Reassign(w http.ResponseWriter, r *http.Request) {
    type reqBody struct {
        PullRequestID string `json:"pull_request_id"`
        OldUserID     string `json:"old_user_id"`
    }
    var req reqBody
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.WriteErrorFromMap(w, model.ErrInvalidInput, http.StatusBadRequest,
            slog.String("path", r.URL.Path))
        return
    }

    if req.PullRequestID == "" || req.OldUserID == "" {
        h.WriteErrorFromMap(w, model.ErrMissingParam, http.StatusBadRequest,
            slog.String("missing_fields", "pull_request_id or old_user_id"))
        return
    }

    pr, replacedBy, err := h.PRRepo.Reassign(r.Context(), req.PullRequestID, req.OldUserID)
    if err != nil {
        status := http.StatusInternalServerError
        if errors.Is(err, model.ErrNotFound) {
            status = http.StatusNotFound
        } else if errors.Is(err, model.ErrPrMerged) {
            status = http.StatusConflict
        } else if errors.Is(err, model.ErrNotAssigned) {
            status = http.StatusConflict
        } else if errors.Is(err, model.ErrNoCandidate) {
            status = http.StatusConflict
        }
        h.WriteErrorFromMap(w, err, status,
            slog.String("pull_request_id", req.PullRequestID),
            slog.String("old_user_id", req.OldUserID))
        return
    }

    h.WriteJSON(w, map[string]any{
        "pr":          pr,
        "replaced_by": replacedBy,
    }, http.StatusOK,
        slog.String("pull_request_id", req.PullRequestID),
        slog.String("replaced_by", replacedBy))
}
