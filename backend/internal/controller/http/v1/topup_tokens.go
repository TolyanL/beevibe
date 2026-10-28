package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) TopupTokens(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, errUnauthorized)

		return
	}

	var body dto.TopupTokensInput

	if !decodeJSON(w, r, &body) {
		writeError(w, errInvalidBody)

		return
	}

	body.UserID = userID
	body.IdempotencyKey = r.Header.Get("Idempotency-Key")

	output, err := h.usecase.User.TopupTokens(r.Context(), body)
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, output)
}
