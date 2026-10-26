package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TolyanL/beevibe/pkg/logger"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logger.Middleware)

	r.Get("/health", h.Health)
	r.Get("/catalog", h.ListCatalog)

	r.Group(func(r chi.Router) {
		r.Use(h.OptionalAuth)
		r.Get("/games/{id}/public", h.GetGamePublicCard)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.RegisterUser)
		r.Post("/login", h.LoginUser)
		r.Post("/logout", h.LogoutUser)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.RequireAuth)

		r.Get("/me", h.GetMe)
		r.Get("/me/token-ops", h.ListTokenOps)
		r.Get("/profile", h.GetProfile)
		r.Patch("/profile", h.UpdateProfile)
		r.Post("/profile/topup", h.TopupTokens)

		r.Get("/games", h.ListUserGames)
		r.Post("/games", h.CreateGame)

		r.Route("/games/{id}", func(r chi.Router) {
			r.Get("/", h.GetGame)
			r.Delete("/", h.DeleteGame)
			r.Patch("/visibility", h.UpdateGameVisibility)
			r.Patch("/meta", h.UpdateGameMeta)
			r.Post("/meta/generate", h.GenerateGameMeta)
			r.Get("/stream", h.StreamGame)
			r.Post("/patch", h.PatchGame)
			r.Post("/cancel", h.CancelGame)
			r.Post("/retry", h.RetryGame)
			r.Patch("/config", h.UpdateGameConfig)
			r.Post("/rollback", h.RollbackGame)
			r.Post("/fix", h.FixGame)
			r.Post("/rate", h.RateGame)
			r.Post("/playtime", h.RecordPlaytime)
		})
	})

	return r
}
