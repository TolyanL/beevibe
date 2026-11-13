package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ListUserGenerationsInput struct {
	UserID uuid.UUID
	Limit  int `json:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset" validate:"omitempty,min=0"`
}

func (i ListUserGenerationsInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type ProfileGeneration struct {
	GameID        uuid.UUID `json:"game_id"`
	GameSlug      string    `json:"game_slug"`
	GameTitle     string    `json:"game_title"`
	GamePrompt    string    `json:"game_prompt"`
	GameStatus    string    `json:"game_status"`
	VersionNumber int       `json:"version_number"`
	Source        string    `json:"source"`
	Note          string    `json:"note"`
	TokensCost    int       `json:"tokens_cost"`
	ModelCalls    int       `json:"model_calls"`
	CreatedAt     time.Time `json:"created_at"`
}

type ListUserGenerationsOutput struct {
	Items []ProfileGeneration `json:"items"`
	Total int                 `json:"total"`
}
