package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type UpdateGameVisibilityInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Visibility string `json:"visibility" validate:"required,oneof=public private"`
}

func (i UpdateGameVisibilityInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type UpdateGameVisibilityOutput struct {
	Visibility string `json:"visibility"`
}
