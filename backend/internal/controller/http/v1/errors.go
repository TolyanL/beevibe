package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/TolyanL/beevibe/internal/domain"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	errInvalidBody  = errors.New("invalid request body")
	errUnauthorized = domain.ErrUnauthorized
)

func writeError(w http.ResponseWriter, err error) {
	code, status := mapError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorBody{Error: errorDetail{Code: code, Message: err.Error()}})
}

func mapError(err error) (string, int) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "not_found", http.StatusNotFound
	case errors.Is(err, domain.ErrVersionNotFound):
		return "version_not_found", http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized):
		return "unauthorized", http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return "forbidden", http.StatusForbidden
	case errors.Is(err, domain.ErrTopupDisabled):
		return "topup_disabled", http.StatusForbidden
	case errors.Is(err, domain.ErrTopupLimitReached):
		return "topup_limit_reached", http.StatusConflict
	case errors.Is(err, domain.ErrNothingToCancel):
		return "nothing_to_cancel", http.StatusConflict
	case errors.Is(err, domain.ErrAlreadyExists):
		return "already_exists", http.StatusConflict
	case errors.Is(err, domain.ErrCannotRateOwnGame):
		return "cannot_rate_own_game", http.StatusForbidden
	case errors.Is(err, domain.ErrEmailTaken):
		return "email_taken", http.StatusConflict
	case errors.Is(err, domain.ErrNotEnoughTokens):
		return "not_enough_tokens", http.StatusConflict
	case errors.Is(err, domain.ErrGenerationInProgress):
		return "generation_in_progress", http.StatusConflict
	case errors.Is(err, domain.ErrInvalidTransition):
		return "invalid_transition", http.StatusConflict
	case errors.Is(err, domain.ErrIdempotencyInProgress):
		return "idempotency_in_progress", http.StatusConflict
	case errors.Is(err, domain.ErrFixLimitReached):
		return "fix_limit_reached", http.StatusConflict
	case errors.Is(err, domain.ErrUnknownConfigKey):
		return "unknown_config_key", http.StatusBadRequest
	case errors.Is(err, domain.ErrValidationFailed):
		return "validation_failed", http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidInput):
		return "invalid_input", http.StatusBadRequest
	default:
		return "invalid_input", http.StatusBadRequest
	}
}
