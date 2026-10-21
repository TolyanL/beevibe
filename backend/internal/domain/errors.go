package domain

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrNotEnoughTokens       = errors.New("not enough tokens")
	ErrInvalidTransition     = errors.New("invalid status transition")
	ErrGenerationInProgress  = errors.New("generation already in progress")
	ErrInvalidInput          = errors.New("invalid input")
	ErrLLMUnavailable        = errors.New("llm unavailable")
	ErrValidationFailed      = errors.New("validation failed")
	ErrIdempotencyInProgress = errors.New("idempotency key already in progress")
	ErrEmailTaken            = errors.New("email already taken")
	ErrVersionNotFound       = errors.New("version not found")
	ErrFixLimitReached       = errors.New("fix limit reached")
	ErrUnknownConfigKey      = errors.New("unknown config key")
	ErrLLMBudgetExceeded     = errors.New("llm call budget exceeded")
)

type Issue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

type ValidationIssues struct {
	Items []Issue
}

func (v *ValidationIssues) Error() string {
	if len(v.Items) == 0 {
		return "validation issues"
	}

	return "validation issues: " + v.Items[0].Message
}
