package dto

import (
	"github.com/google/uuid"
)

type CancelGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID
}

type CancelGameOutput struct {
	Status string `json:"status"`
}
