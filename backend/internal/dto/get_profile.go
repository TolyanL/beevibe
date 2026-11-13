package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

type ProfileGame struct {
	ID          uuid.UUID         `json:"id"`
	Prompt      string            `json:"prompt"`
	Genre       string            `json:"genre"`
	Status      string            `json:"status"`
	Visibility  string            `json:"visibility"`
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	Description string            `json:"description"`
	Speed       string            `json:"speed"`
	Cover       *domain.GameCover `json:"cover,omitempty"`
	TokensCost  int               `json:"tokens_cost"`
	ModelCalls  int               `json:"model_calls"`
	Fixes       int               `json:"fixes"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type ProfileTokenOp struct {
	GameID    *uuid.UUID `json:"game_id,omitempty"`
	GameSlug  *string    `json:"game_slug,omitempty"`
	GameLabel string     `json:"game_label,omitempty"`
	Delta     int64      `json:"delta"`
	Reason    string     `json:"reason"`
	CreatedAt time.Time  `json:"created_at"`
}

type ProfilePrices struct {
	Generate     int64 `json:"generate"`
	GenerateFast int64 `json:"generate_fast"`
	Patch        int64 `json:"patch"`
	Meta         int64 `json:"meta"`
}

type ProfileStats struct {
	GamesTotal  int64 `json:"games_total"`
	GamesReady  int64 `json:"games_ready"`
	TokensSpent int64 `json:"tokens_spent"`
	Fixes       int64 `json:"fixes"`
}

type GetProfileInput struct {
	UserID uuid.UUID
}

type GetProfileOutput struct {
	Balance  int64            `json:"balance"`
	Stats    ProfileStats     `json:"stats"`
	Games    []ProfileGame    `json:"games"`
	TokenOps []ProfileTokenOp `json:"token_ops"`
	Prices   ProfilePrices    `json:"prices"`
}
