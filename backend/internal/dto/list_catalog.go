package dto

import (
	"fmt"
)

type ListCatalogInput struct {
	Sort   string `json:"sort"   validate:"omitempty,oneof=rating hours fresh"`
	Genre  string `json:"genre"  validate:"omitempty,oneof=arcade snake clicker platformer puzzle custom"`
	Limit  int    `json:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int    `json:"offset" validate:"omitempty,min=0"`
}

func (i ListCatalogInput) Validate() error {
	if err := validateStruct(i); err != nil {
		return fmt.Errorf("validateStruct: %w", err)
	}

	return nil
}

type ListCatalogOutput struct {
	Items []CatalogCard `json:"items"`
}
