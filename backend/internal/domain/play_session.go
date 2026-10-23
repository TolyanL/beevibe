package domain

import (
	"time"

	"github.com/google/uuid"
)

const MaxPlaytimeSeconds = 300

type PlaySession struct {
	ID        uuid.UUID  `json:"id"`
	GameID    uuid.UUID  `json:"game_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Seconds   int        `json:"seconds"`
	CreatedAt time.Time  `json:"created_at"`
}

func ClampPlaytime(seconds int) int {
	if seconds < 0 {
		return 0
	}

	if seconds > MaxPlaytimeSeconds {
		return MaxPlaytimeSeconds
	}

	return seconds
}
