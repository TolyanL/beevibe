package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) LoginUser(w http.ResponseWriter, r *http.Request) {
	var body dto.LoginUserInput

	if !decodeJSON(w, r, &body) {
		writeError(w, errInvalidBody)

		return
	}

	output, err := h.usecase.User.LoginUser(r.Context(), body)
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	setSessionCookie(w, output.SessionToken)
	writeJSON(w, http.StatusOK, output)
}
