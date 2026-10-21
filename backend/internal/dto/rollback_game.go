package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type RollbackGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Version int `json:"version" validate:"required,min=1"`
}

func (i RollbackGameInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type RollbackGameOutput struct {
	VersionID uuid.UUID `json:"versionId"`
}
