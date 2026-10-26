package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type RateGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Value int `json:"value" validate:"required,min=1,max=5"`
}

func (i RateGameInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type RateGameOutput struct {
	Rating    int             `json:"rating"`
	Aggregate RatingAggregate `json:"aggregate"`
}
