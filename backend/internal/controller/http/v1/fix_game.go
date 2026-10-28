package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) FixGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, errUnauthorized)

		return
	}

	gameID, ok := gameIDFromContext(r.Context())
	if !ok {
		writeError(w, errInvalidBody)

		return
	}

	var body dto.FixGameInput

	if !decodeJSON(w, r, &body) {
		writeError(w, errInvalidBody)

		return
	}

	body.UserID = userID
	body.GameID = gameID
	body.IdempotencyKey = r.Header.Get("Idempotency-Key")

	output, err := h.usecase.Game.Fix(r.Context(), body)
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusAccepted, output)
}
