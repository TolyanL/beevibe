package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type RateGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Value string `json:"value" validate:"required,oneof=cool bad"`
}

func (i RateGameInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type RateGameOutput struct {
	Rating    string          `json:"rating"`
	Aggregate RatingAggregate `json:"aggregate"`
}
