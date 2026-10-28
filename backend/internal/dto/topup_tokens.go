package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type TopupTokensInput struct {
	UserID uuid.UUID

	Amount         int64  `json:"amount"    validate:"required,min=1"`
	IdempotencyKey string `json:"-" `
}

func (i TopupTokensInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type TopupTokensOutput struct {
	Balance int64 `json:"balance"`
}
