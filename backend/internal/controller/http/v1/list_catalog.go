package v1

import (
	"net/http"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/render"
)

func (h *Handlers) ListCatalog(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	input := dto.ListCatalogInput{
		Sort:   query.Get("sort"),
		Genre:  query.Get("genre"),
		Limit:  atoiDefault(query.Get("limit"), 0),
		Offset: atoiDefault(query.Get("offset"), 0),
	}

	output, err := h.usecase.ListCatalog(r.Context(), input)
	if err != nil {
		err = render.Error(r.Context(), err, "request failed")

		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, output)
}
