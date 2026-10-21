package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) LogoutUser(w http.ResponseWriter, r *http.Request) {
	var input dto.LogoutUserInput

	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		input.SessionToken = cookie.Value
	}

	if _, err := h.usecase.LogoutUser(r.Context(), input); err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
