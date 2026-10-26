package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TokenReason string

const (
	ReasonSignupGrant TokenReason = "signup_grant"
	ReasonGenerate    TokenReason = "generate"
	ReasonPatch       TokenReason = "patch"
	ReasonMeta        TokenReason = "meta"
	ReasonTopup       TokenReason = "topup"
	ReasonRefund      TokenReason = "refund"
)

type TokenOp struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	GameID    *uuid.UUID  `json:"game_id,omitempty"`
	Delta     int64       `json:"delta"`
	Reason    TokenReason `json:"reason"`
	CreatedAt time.Time   `json:"created_at"`
}

func NewTokenOp(userID uuid.UUID, gameID *uuid.UUID, delta int64, reason TokenReason) (TokenOp, error) {
	if delta == 0 {
		return TokenOp{}, fmt.Errorf("%w: token delta must not be zero", ErrInvalidInput)
	}

	return TokenOp{
		ID:        uuid.New(),
		UserID:    userID,
		GameID:    gameID,
		Delta:     delta,
		Reason:    reason,
		CreatedAt: time.Now().UTC(),
	}, nil
}
