package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type ResolveReviewInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Action string `json:"action" validate:"required,oneof=accept regenerate"`
}

func (i ResolveReviewInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type ResolveReviewOutput struct {
	Status     string `json:"status"`
	ResumeFrom string `json:"resume_from,omitempty"`
}
