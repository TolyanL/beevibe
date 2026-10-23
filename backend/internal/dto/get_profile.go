package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProfileGame struct {
	ID         uuid.UUID `json:"id"`
	Prompt     string    `json:"prompt"`
	Genre      string    `json:"genre"`
	Status     string    `json:"status"`
	TokensCost int       `json:"tokens_cost"`
	ModelCalls int       `json:"model_calls"`
	Fixes      int       `json:"fixes"`
	CreatedAt  time.Time `json:"created_at"`
}

type ProfileTokenOp struct {
	GameID    *uuid.UUID `json:"game_id,omitempty"`
	Delta     int64      `json:"delta"`
	Reason    string     `json:"reason"`
	CreatedAt time.Time  `json:"created_at"`
}

type ProfilePrices struct {
	Generate int64 `json:"generate"`
	Patch    int64 `json:"patch"`
}

type GetProfileInput struct {
	UserID uuid.UUID
}

type GetProfileOutput struct {
	Balance  int64            `json:"balance"`
	Games    []ProfileGame    `json:"games"`
	TokenOps []ProfileTokenOp `json:"token_ops"`
	Prices   ProfilePrices    `json:"prices"`
}
