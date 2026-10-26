package dto

import (
	"github.com/google/uuid"
)

type DeleteGameInput struct {
	UserID uuid.UUID
	GameID uuid.UUID
}

type DeleteGameOutput struct {
	Deleted bool `json:"deleted"`
}
