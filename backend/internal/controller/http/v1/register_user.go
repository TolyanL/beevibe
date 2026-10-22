package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var body dto.RegisterUserInput

	if !decodeJSON(w, r, &body) {
		writeError(w, errInvalidBody)

		return
	}

	output, err := h.usecase.User.RegisterUser(r.Context(), body)
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	setSessionCookie(w, output.SessionToken)
	writeJSON(w, http.StatusCreated, output)
}
