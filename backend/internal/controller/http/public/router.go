package public

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/logger"
)

func NewRouter(h *Handlers, spa http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logger.Middleware)

	if spa != nil {
		r.NotFound(spa.ServeHTTP)
	}

	r.Group(func(r chi.Router) {
		r.Use(h.OptionalAuth)

		r.Get("/g/{id}", h.GetGame)
		r.Get("/g/{id}/download", h.DownloadGame)
	})

	return r
}

func (h *Handlers) GetGame(w http.ResponseWriter, r *http.Request) {
	output, err := h.usecase.Game.GetPublic(r.Context(), publicInput(r, parseGameSlug(r)))
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "game not found")

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output.HTML))
}

func (h *Handlers) DownloadGame(w http.ResponseWriter, r *http.Request) {
	output, err := h.usecase.Game.GetPublic(r.Context(), publicInput(r, parseGameSlug(r)))
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "game not found")

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="game.html"`)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output.HTML))
}

func publicInput(r *http.Request, slug string) dto.GetGamePublicInput {
	input := dto.GetGamePublicInput{Slug: slug}

	if viewerID, ok := viewerIDFromContext(r.Context()); ok {
		input.ViewerID = &viewerID
	}

	return input
}
