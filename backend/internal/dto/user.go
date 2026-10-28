package dto

import "github.com/google/uuid"

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Balance     int64     `json:"balance"`
}
