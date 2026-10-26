package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RatingValue int

const (
	Rating1 RatingValue = 1
	Rating2 RatingValue = 2
	Rating3 RatingValue = 3
	Rating4 RatingValue = 4
	Rating5 RatingValue = 5
)

func (v RatingValue) Valid() bool {
	return v >= Rating1 && v <= Rating5
}

type Rating struct {
	GameID    uuid.UUID   `json:"game_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Value     RatingValue `json:"value"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func NewRating(gameID, userID uuid.UUID, value RatingValue) (Rating, error) {
	if !value.Valid() {
		return Rating{}, fmt.Errorf("%w: rating value must be between 1 and 5", ErrInvalidInput)
	}

	return Rating{
		GameID:    gameID,
		UserID:    userID,
		Value:     value,
		UpdatedAt: time.Now().UTC(),
	}, nil
}
