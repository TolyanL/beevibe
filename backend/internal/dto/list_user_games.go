package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type ListUserGamesInput struct {
	UserID uuid.UUID
	Status string `json:"status" validate:"omitempty,oneof=queued analyzing planning coding assets reviewing assembling checking patching ready failed"`
	Sort   string `json:"sort"   validate:"omitempty,oneof=created_at tokens_cost"`
	Limit  int    `json:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int    `json:"offset" validate:"omitempty,min=0"`
}

func (i ListUserGamesInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type ListUserGamesOutput struct {
	Items []ProfileGame `json:"items"`
	Total int           `json:"total"`
}
