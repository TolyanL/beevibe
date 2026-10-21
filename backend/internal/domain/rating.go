package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RatingValue string

const (
	RatingCool RatingValue = "cool"
	RatingBad  RatingValue = "bad"
)

type Rating struct {
	GameID    uuid.UUID   `json:"game_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Value     RatingValue `json:"value"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func NewRating(gameID, userID uuid.UUID, value RatingValue) (Rating, error) {
	if value != RatingCool && value != RatingBad {
		return Rating{}, fmt.Errorf("%w: rating value must be cool or bad", ErrInvalidInput)
	}

	return Rating{
		GameID:    gameID,
		UserID:    userID,
		Value:     value,
		UpdatedAt: time.Now().UTC(),
	}, nil
}
