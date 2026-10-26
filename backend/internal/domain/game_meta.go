package domain

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	TitleMaxLength       = 80
	DescriptionMaxLength = 500
	coverPaletteSize     = 5
)

type GameCover struct {
	Palette []string `json:"palette"`
	Style   string   `json:"style"`
	Seed    string   `json:"seed"`
}

func (m GameMeta) Validate() error {
	if len([]rune(m.Title)) > TitleMaxLength {
		return fmt.Errorf("%w: title is longer than %d characters", ErrInvalidInput, TitleMaxLength)
	}

	if len([]rune(m.Description)) > DescriptionMaxLength {
		return fmt.Errorf("%w: description is longer than %d characters", ErrInvalidInput, DescriptionMaxLength)
	}

	if hasControlRunes(m.Title) || hasControlRunes(m.Description) {
		return fmt.Errorf("%w: title and description must not contain control characters", ErrInvalidInput)
	}

	if m.Cover != nil {
		if err := m.Cover.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type GameMeta struct {
	Title       string
	Description string
	Cover       *GameCover
}

func NewGameMeta(title, description string, cover *GameCover) (GameMeta, error) {
	meta := GameMeta{
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Cover:       cover,
	}

	if err := meta.Validate(); err != nil {
		return GameMeta{}, fmt.Errorf("meta.Validate: %w", err)
	}

	return meta, nil
}

func (c GameCover) Validate() error {
	if len(c.Palette) != 0 && len(c.Palette) != coverPaletteSize {
		return fmt.Errorf("%w: cover palette must contain %d colors", ErrInvalidInput, coverPaletteSize)
	}

	for _, color := range c.Palette {
		if !isHexColor(color) {
			return fmt.Errorf("%w: cover color %q is not a hex color", ErrInvalidInput, color)
		}
	}

	if len([]rune(c.Style)) > 32 || len([]rune(c.Seed)) > 32 {
		return fmt.Errorf("%w: cover style and seed must be shorter than 32 characters", ErrInvalidInput)
	}

	if hasControlRunes(c.Style) || hasControlRunes(c.Seed) {
		return fmt.Errorf("%w: cover style and seed must not contain control characters", ErrInvalidInput)
	}

	return nil
}

func hasControlRunes(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return true
		}
	}

	return false
}

func isHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}

	for _, r := range color[1:] {
		isDigit := r >= '0' && r <= '9'
		isLowerHex := r >= 'a' && r <= 'f'
		isUpperHex := r >= 'A' && r <= 'F'

		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
	}

	return true
}
