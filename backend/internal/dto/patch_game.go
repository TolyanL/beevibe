package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type PatchGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Text           string `json:"text" validate:"required,min=2,max=4000"`
	IdempotencyKey string `json:"-"`
}

func (i PatchGameInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type PatchGameOutput struct {
	Status string `json:"status"`
}
