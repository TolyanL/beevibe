package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

type Stage struct {
	Stage      string `json:"stage"`
	Status     string `json:"status"`
	DurationMs int64  `json:"durationMs"`
	Issues     int    `json:"issues,omitempty"`
}

type Version struct {
	ID            uuid.UUID `json:"id"`
	Number        int       `json:"number"`
	Source        string    `json:"source"`
	Note          string    `json:"note,omitempty"`
	TokensCost    int       `json:"tokens_cost"`
	ModelCalls    int       `json:"model_calls"`
	PromptVersion string    `json:"prompt_version,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type GameCost struct {
	Tokens int `json:"tokens"`
	Calls  int `json:"calls"`
	Fixes  int `json:"fixes"`
}

type GetGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID
}

type GetGameOutput struct {
	Status   string             `json:"status"`
	Stages   []Stage            `json:"stages"`
	Spec     *domain.GameSpec   `json:"spec,omitempty"`
	Config   *domain.GameConfig `json:"config,omitempty"`
	Versions []Version          `json:"versions"`
	Cost     GameCost           `json:"cost"`
	Rating   RatingAggregate    `json:"rating"`
	MyRating int                `json:"my_rating"`
}
