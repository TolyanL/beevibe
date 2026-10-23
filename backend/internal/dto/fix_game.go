package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type FixGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Error string `json:"error" validate:"omitempty,max=2000"`
}

func (i FixGameInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type FixGameOutput struct {
	VersionID uuid.UUID `json:"versionId"`
}
