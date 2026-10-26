package public

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func parseGameRef(r *http.Request) (uuid.UUID, string) {
	raw := chi.URLParam(r, "id")

	if gameID, err := uuid.Parse(raw); err == nil {
		return gameID, ""
	}

	return uuid.Nil, raw
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	//nolint:errcheck
	json.NewEncoder(w).Encode(errorBody{Error: errorDetail{Code: code, Message: message}})
}
