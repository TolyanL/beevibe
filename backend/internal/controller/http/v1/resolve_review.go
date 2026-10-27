package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) AcceptReview(w http.ResponseWriter, r *http.Request) {
	h.resolveReview(w, r, "accept")
}

func (h *Handlers) RegenerateReview(w http.ResponseWriter, r *http.Request) {
	h.resolveReview(w, r, "regenerate")
}

func (h *Handlers) resolveReview(w http.ResponseWriter, r *http.Request, action string) {
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

	output, err := h.usecase.Game.ResolveReview(r.Context(), dto.ResolveReviewInput{
		UserID: userID,
		GameID: gameID,
		Action: action,
	})
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusAccepted, output)
}
