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
	Operation  string `json:"operation,omitempty"`
}

type ActiveOperation struct {
	Kind      string    `json:"kind"`
	StartedAt time.Time `json:"started_at"`
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
	ID           uuid.UUID          `json:"id"`
	Prompt       string             `json:"prompt"`
	Params       domain.GameParams  `json:"params"`
	Status       string             `json:"status"`
	FailReason   string             `json:"fail_reason,omitempty"`
	Visibility   string             `json:"visibility"`
	Title        string             `json:"title"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description"`
	Cover        *domain.GameCover  `json:"cover,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	Stages       []Stage            `json:"stages"`
	Spec         *domain.GameSpec   `json:"spec,omitempty"`
	Config       *domain.GameConfig `json:"config,omitempty"`
	Versions     []Version          `json:"versions"`
	Cost         GameCost           `json:"cost"`
	Rating       RatingAggregate    `json:"rating"`
	MyRating     int                `json:"my_rating"`
	Operation    *ActiveOperation   `json:"operation,omitempty"`
	ReviewIssues []domain.Issue     `json:"review_issues,omitempty"`
	Cursor       int64              `json:"cursor"`
}
