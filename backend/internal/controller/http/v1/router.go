package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TolyanL/beevibe/pkg/logger"
)

// NewRouter — ручки API. Монтируется в /api (см. internal/app).
func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logger.Middleware)

	r.Get("/health", h.Health)
	r.Get("/catalog", h.ListCatalog)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.RegisterUser)
		r.Post("/login", h.LoginUser)
		r.Post("/logout", h.LogoutUser)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.RequireAuth)

		r.Get("/me", h.GetMe)
		r.Get("/profile", h.GetProfile)
		r.Post("/profile/topup", h.TopupTokens)

		r.Post("/games", h.CreateGame)

		r.Route("/games/{id}", func(r chi.Router) {
			r.Get("/", h.GetGame)
			r.Get("/stream", h.StreamGame)
			r.Post("/patch", h.PatchGame)
			r.Patch("/config", h.UpdateGameConfig)
			r.Post("/rollback", h.RollbackGame)
			r.Post("/fix", h.FixGame)
			r.Post("/rate", h.RateGame)
			r.Post("/playtime", h.RecordPlaytime)
		})
	})

	return r
}
