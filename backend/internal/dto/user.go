package dto

import "github.com/google/uuid"

type User struct {
	ID      uuid.UUID `json:"id"`
	Email   string    `json:"email"`
	Balance int64     `json:"balance"`
}
