package handlers

import (
	"encoding/json"
	"errors"
	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	repository "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository"
	"log/slog"
	"net/http"
)

type UserHandler struct {
	BaseHandler
	UserRepo repository.UserRepository
}

func NewUserHandler(logger slog.Logger, userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		BaseHandler: BaseHandler{Logger: logger},
		UserRepo:    userRepo,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	var req reqBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorFromMap(w, model.ErrInvalidInput, http.StatusBadRequest,
			slog.String("path", r.URL.Path))
		return
	}

	user, err := h.UserRepo.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotFound) {
			status = http.StatusNotFound
		}
		h.WriteErrorFromMap(w, err, status, slog.String("user_id", req.UserID))
		return
	}
	h.WriteJSON(w, map[string]any{"user": user}, http.StatusOK, slog.String("user_id", req.UserID))
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.WriteErrorFromMap(w, model.ErrMissingParam, http.StatusBadRequest,
			slog.String("query", r.URL.RawQuery))
		return
	}

	prs, err := h.UserRepo.GetReview(r.Context(), userID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotFound) {
			status = http.StatusNotFound
		}
		h.WriteErrorFromMap(w, err, status, slog.String("user_id", userID))
		return
	}
	h.WriteJSON(w, map[string]any{
		"user_id":       userID,
		"pull_requests": prs,
	}, http.StatusOK, slog.String("user_id", userID))
}
