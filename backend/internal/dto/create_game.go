package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type CreateGameInput struct {
	UserID uuid.UUID

	Prompt  string `json:"prompt"  validate:"required,min=3,max=2000"`
	Genre   string `json:"genre"   validate:"omitempty,oneof=arcade snake clicker platformer puzzle custom"`
	Tempo   *int   `json:"tempo"   validate:"omitempty,min=1,max=10"`
	MapSize string `json:"mapSize" validate:"omitempty,oneof=small medium large"`
	Style   string `json:"style"   validate:"omitempty,oneof=pixel cartoon neon minimal"`

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
	Status string    `json:"status"`
}
