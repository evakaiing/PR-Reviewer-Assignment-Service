package handlers

import (
	"encoding/json"
	"errors"
	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	repository "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository"
	"log/slog"
	"net/http"
)

type TeamHandler struct {
	BaseHandler
	TeamRepo repository.TeamRepository
}

func NewTeamHandler(logger slog.Logger, teamRepo repository.TeamRepository) *TeamHandler {
	return &TeamHandler{
		BaseHandler: BaseHandler{Logger: logger},
		TeamRepo:    teamRepo,
	}
}

func (h *TeamHandler) Add(w http.ResponseWriter, r *http.Request) {
	var team model.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		h.WriteErrorFromMap(w, model.ErrInvalidInput, http.StatusBadRequest,
			slog.String("path", r.URL.Path))
		return
	}

	if team.TeamName == "" {
		h.WriteErrorFromMap(w, model.ErrMissingParam, http.StatusBadRequest,
			slog.String("field", "team_name"))
		return
	}

	err := h.TeamRepo.Add(r.Context(), &team)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrTeamExists) {
			status = http.StatusBadRequest
		}
		h.WriteErrorFromMap(w, err, status, slog.String("team_name", team.TeamName))
		return
	}

	createdTeam, err := h.TeamRepo.Get(r.Context(), team.TeamName)
	if err != nil {
		h.WriteErrorFromMap(w, err, http.StatusInternalServerError,
			slog.String("team_name", team.TeamName))
		return
	}

	h.WriteJSON(w, map[string]any{"team": createdTeam}, http.StatusCreated,
		slog.String("team_name", team.TeamName))
}

func (h *TeamHandler) Get(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.WriteErrorFromMap(w, model.ErrMissingParam, http.StatusBadRequest,
			slog.String("query", r.URL.RawQuery))
		return
	}

	team, err := h.TeamRepo.Get(r.Context(), teamName)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotFound) {
			status = http.StatusNotFound
		}
		h.WriteErrorFromMap(w, err, status, slog.String("team_name", teamName))
		return
	}

	h.WriteJSON(w, team, http.StatusOK, slog.String("team_name", teamName))
}
