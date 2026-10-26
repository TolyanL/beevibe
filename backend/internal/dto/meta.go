package dto

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

type UpdateGameMetaInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	Title       *string           `json:"title"       validate:"omitempty,max=80"`
	Description *string           `json:"description" validate:"omitempty,max=500"`
	Cover       *domain.GameCover `json:"cover"`
}

func (i UpdateGameMetaInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type UpdateGameMetaOutput struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Cover       *domain.GameCover `json:"cover,omitempty"`
}

type GenerateGameMetaInput struct {
	UserID uuid.UUID
	GameID uuid.UUID

	IdempotencyKey string `json:"-"`
}

type GenerateGameMetaOutput struct {
	Status string `json:"status"`
}
