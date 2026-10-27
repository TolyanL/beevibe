package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) DeleteGame(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.usecase.Game.Delete(r.Context(), dto.DeleteGameInput{UserID: userID, GameID: gameID})
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, output)
}
