package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) GetGame(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.usecase.Game.Get(r.Context(), dto.GetGameInput{UserID: userID, GameID: gameID})
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, output)
}
