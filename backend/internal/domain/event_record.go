package domain

import (
	"time"

	"github.com/google/uuid"
)

// GameEventRecord — запись события, полученного из Kafka (проекция game_event_log).
type GameEventRecord struct {
	ID         uuid.UUID `json:"id"`
	EventID    uuid.UUID `json:"event_id"`
	Topic      string    `json:"topic"`
	Key        string    `json:"key"`
	EventType  EventType `json:"event_type"`
	Payload    []byte    `json:"payload"`
	ReceivedAt time.Time `json:"received_at"`
}
