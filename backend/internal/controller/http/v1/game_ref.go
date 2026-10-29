package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

type gameIDKey struct{}

func gameIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	gameID, ok := ctx.Value(gameIDKey{}).(uuid.UUID)

	return gameID, ok
}

func (h *Handlers) ResolveGame(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ref := chi.URLParam(r, "id")

		gameID, err := h.usecase.Game.ResolveGameID(r.Context(), ref)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				writeError(w, domain.ErrNotFound)

				return
			}

			writeError(w, errInvalidBody)

			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), gameIDKey{}, gameID)))
	})
}
