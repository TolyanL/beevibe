package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventGameCreated EventType = "game.created"
	EventGameReady   EventType = "game.ready"
	EventGameFailed  EventType = "game.failed"
	EventGamePatched EventType = "game.patched"
)

const GameEventsTopic = "game.events"

type Event struct {
	ID        uuid.UUID `json:"id"`
	Type      EventType `json:"type"`
	Topic     string    `json:"topic"`
	Key       string    `json:"key"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type GameEventPayload struct {
	ID         uuid.UUID  `json:"id"`
	Type       EventType  `json:"type"`
	GameID     uuid.UUID  `json:"game_id"`
	OwnerID    uuid.UUID  `json:"owner_id"`
	Prompt     string     `json:"prompt"`
	Status     Status     `json:"status"`
	VersionID  *uuid.UUID `json:"version_id,omitempty"`
	TokensCost int        `json:"tokens_cost"`
	Error      string     `json:"error,omitempty"`
	OccurredAt time.Time  `json:"occurred_at"`
}
