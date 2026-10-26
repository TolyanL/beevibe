package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type UpdateProfileInput struct {
	UserID uuid.UUID

	DisplayName string `json:"display_name" validate:"max=40"`
}

func (i UpdateProfileInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type UpdateProfileOutput struct {
	User User `json:"user"`
}
