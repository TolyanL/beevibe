package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) ListUserGenerations(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, errUnauthorized)

		return
	}

	query := r.URL.Query()

	output, err := h.usecase.Game.ListUserGenerations(r.Context(), dto.ListUserGenerationsInput{
		UserID: userID,
		Limit:  atoiDefault(query.Get("limit"), 0),
		Offset: atoiDefault(query.Get("offset"), 0),
	})
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, output)
}
