package dto

import (
	"github.com/google/uuid"
)

type RetryGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID
}

type RetryGameOutput struct {
	Status string `json:"status"`
}
