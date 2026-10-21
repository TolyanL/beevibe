package dto

import (
	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

type ListGenEventsInput struct {
	UserID   uuid.UUID
	GameID   uuid.UUID
	AfterSeq int64
}

type ListGenEventsOutput struct {
	Events []domain.GenEvent `json:"events"`
	Status string            `json:"status"`
}
