package dto

import "github.com/google/uuid"

type GetMeInput struct {
	UserID uuid.UUID
}

type GetMeOutput struct {
	User User `json:"user"`
}
