package dto

import "fmt"

type RegisterUserInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

func (i RegisterUserInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type RegisterUserOutput struct {
	User User `json:"user"`

	SessionToken string `json:"-"`
}
