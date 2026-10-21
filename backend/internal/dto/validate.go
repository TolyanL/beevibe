package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled()) //nolint:gochecknoglobals

func validateStruct(input any) error {
	if err := validate.Struct(input); err != nil {
		return fmt.Errorf("validate.Struct: %w", err)
	}

	return nil
}
