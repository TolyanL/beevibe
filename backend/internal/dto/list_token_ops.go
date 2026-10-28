package dto

import (
	"fmt"
)

type ListTokenOpsInput struct {
	Limit  int `json:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset" validate:"omitempty,min=0"`
}

func (i ListTokenOpsInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type ListTokenOpsOutput struct {
	Items []ProfileTokenOp `json:"items"`
	Total int              `json:"total"`
}
