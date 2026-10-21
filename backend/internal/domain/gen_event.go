package domain

import (
	"time"

	"github.com/google/uuid"
)

type GenEventType string

const (
	GenEventStage    GenEventType = "stage"
	GenEventArtifact GenEventType = "artifact"
	GenEventDone     GenEventType = "done"
	GenEventError    GenEventType = "error"
)

type GenEvent struct {
	GameID    uuid.UUID    `json:"game_id"`
	Seq       int64        `json:"seq"`
	Type      GenEventType `json:"type"`
	Payload   []byte       `json:"payload"`
	CreatedAt time.Time    `json:"created_at"`
}
