package dto

import (
	"github.com/google/uuid"
)

type GetGamePublicInput struct {
	GameID uuid.UUID
	Slug   string
}

type GetGamePublicOutput struct {
	HTML string `json:"html"`
}
