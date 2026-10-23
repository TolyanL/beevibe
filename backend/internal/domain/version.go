package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type VersionSource string

const (
	VersionSourceGenerate VersionSource = "generate"
	VersionSourceReview   VersionSource = "review"
	VersionSourcePatch    VersionSource = "patch"
	VersionSourceConfig   VersionSource = "config"
	VersionSourceRepair   VersionSource = "repair"
	VersionSourceRollback VersionSource = "rollback"
)

type Version struct {
	ID            uuid.UUID     `json:"id"`
	GameID        uuid.UUID     `json:"game_id"`
	Number        int           `json:"number"`
	HTML          string        `json:"html"`
	Source        VersionSource `json:"source"`
	Note          string        `json:"note,omitempty"`
	TokensCost    int           `json:"tokens_cost"`
	ModelCalls    int           `json:"model_calls"`
	PromptVersion string        `json:"prompt_version,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

func NewVersion(gameID uuid.UUID, number int, html string, source VersionSource) (Version, error) {
	if number <= 0 {
		return Version{}, fmt.Errorf("%w: version number must be positive", ErrInvalidInput)
	}

	if html == "" {
		return Version{}, fmt.Errorf("%w: html is required", ErrInvalidInput)
	}

	return Version{
		ID:        uuid.New(),
		GameID:    gameID,
		Number:    number,
		HTML:      html,
		Source:    source,
		CreatedAt: time.Now().UTC(),
	}, nil
}
