package public

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func parseGameSlug(r *http.Request) string {
	return chi.URLParam(r, "id")
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorBody{Error: errorDetail{Code: code, Message: message}})
}
