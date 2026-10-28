package dto

import (
	"github.com/google/uuid"
)

type GetGamePublicInput struct {
	Slug     string
	ViewerID *uuid.UUID
}

type GetGamePublicOutput struct {
	HTML string `json:"html"`
}
