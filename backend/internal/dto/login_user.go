package dto

import "fmt"

type LoginUserInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (i LoginUserInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type LoginUserOutput struct {
	User User `json:"user"`

	SessionToken string `json:"-"`
}
