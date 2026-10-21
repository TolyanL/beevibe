package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type UpdateGameConfigInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Tempo      *int     `json:"tempo"      validate:"omitempty,min=1,max=10"`
	Palette    []string `json:"palette"    validate:"omitempty,max=8,dive,required"`
	MapSize    *string  `json:"mapSize"    validate:"omitempty,oneof=small medium large"`
	Difficulty *string  `json:"difficulty" validate:"omitempty,max=32"`
}

func (i UpdateGameConfigInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type UpdateGameConfigOutput struct {
	VersionID uuid.UUID `json:"versionId"`
}
