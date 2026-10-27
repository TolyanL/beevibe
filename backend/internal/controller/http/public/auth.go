package public

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const sessionCookieName = "session"

type viewerIDKey struct{}

func viewerIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	viewerID, ok := ctx.Value(viewerIDKey{}).(uuid.UUID)

	return viewerID, ok
}

// OptionalAuth пропускает зрителя дальше без сессии: приватную игру
// показываем только её автору, для остальных она остаётся невидимой.
func (h *Handlers) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)

			return
		}

		viewerID, err := h.usecase.User.GetSessionUserID(r.Context(), cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)

			return
		}

		ctx := context.WithValue(r.Context(), viewerIDKey{}, viewerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
