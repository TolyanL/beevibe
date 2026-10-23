package dto

import "github.com/google/uuid"

type TopupTokensInput struct {
	UserID uuid.UUID
}

type TopupTokensOutput struct {
	Balance int64 `json:"balance"`
}
