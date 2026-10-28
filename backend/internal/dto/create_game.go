package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type CreateGameInput struct {
	UserID uuid.UUID

	Prompt  string `json:"prompt"  validate:"required,min=3,max=2000"`
	Title   string `json:"title"   validate:"omitempty,max=80"`
	Genre   string `json:"genre"   validate:"omitempty,oneof=arcade snake clicker platformer puzzle custom"`
	Tempo   *int   `json:"tempo"   validate:"omitempty,min=1,max=10"`
	MapSize string `json:"mapSize" validate:"omitempty,oneof=small medium large"`
	Style   string `json:"style"   validate:"omitempty,oneof=pixel cartoon neon minimal custom"`

	CustomGenre string `json:"customGenre" validate:"required_if=Genre custom,omitempty,min=2,max=32"`
	CustomStyle string `json:"customStyle" validate:"required_if=Style custom,omitempty,min=2,max=160"`

	IdempotencyKey string `json:"-"`
}

func (i CreateGameInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type CreateGameOutput struct {
	ID     uuid.UUID `json:"id"`
	Slug   string    `json:"slug"`
	Status string    `json:"status"`
}
