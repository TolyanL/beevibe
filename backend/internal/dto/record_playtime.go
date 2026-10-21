package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type RecordPlaytimeInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Seconds int `json:"seconds" validate:"required,min=0,max=300"`
}

func (i RecordPlaytimeInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type RecordPlaytimeOutput struct {
	TotalSeconds int `json:"totalSeconds"`
}
