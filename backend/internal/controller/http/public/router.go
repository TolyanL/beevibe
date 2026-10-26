package public

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/logger"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logger.Middleware)

	r.Get("/g/{id}", h.GetGame)
	r.Get("/g/{id}/download", h.DownloadGame)

	return r
}

func (h *Handlers) GetGame(w http.ResponseWriter, r *http.Request) {
	gameID, slug := parseGameRef(r)

	output, err := h.usecase.Game.GetPublic(r.Context(), dto.GetGamePublicInput{GameID: gameID, Slug: slug})
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "game not found")

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	//nolint:errcheck
	w.Write([]byte(output.HTML))
}

func (h *Handlers) DownloadGame(w http.ResponseWriter, r *http.Request) {
	gameID, slug := parseGameRef(r)

	output, err := h.usecase.Game.GetPublic(r.Context(), dto.GetGamePublicInput{GameID: gameID, Slug: slug})
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "game not found")

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="game.html"`)
	w.WriteHeader(http.StatusOK)

	//nolint:errcheck
	w.Write([]byte(output.HTML))
}
