package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) GetGamePublicCard(w http.ResponseWriter, r *http.Request) {
	var input dto.GetGamePublicCardInput

	input.Slug = chi.URLParam(r, "id")

	viewerID, ok := userIDFromContext(r.Context())
	if ok {
		input.ViewerID = &viewerID
	}

	output, err := h.usecase.Game.GetPublicCard(r.Context(), input)
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, output)
}
