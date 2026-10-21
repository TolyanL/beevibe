package domain

import (
	"fmt"
)

type GameConfig struct {
	Tempo      Tempo    `json:"tempo"`
	Palette    []string `json:"palette"`
	MapSize    MapSize  `json:"map_size"`
	Difficulty string   `json:"difficulty"`
	Controls   string   `json:"controls"`
	Style      Style    `json:"style"`
}

type GameConfigUpdate struct {
	Tempo      *Tempo    `json:"tempo,omitempty"`
	Palette    *[]string `json:"palette,omitempty"`
	MapSize    *MapSize  `json:"map_size,omitempty"`
	Difficulty *string   `json:"difficulty,omitempty"`
	Style      *Style    `json:"style,omitempty"`
}

func (c GameConfig) Validate() error {
	if c.Tempo < TempoSlow || c.Tempo > TempoFast {
		return fmt.Errorf("%w: tempo must be between %d and %d", ErrInvalidInput, TempoSlow, TempoFast)
	}

	switch c.MapSize {
	case MapSizeSmall, MapSizeMedium, MapSizeLarge:
	default:
		return fmt.Errorf("%w: unknown map size %q", ErrInvalidInput, c.MapSize)
	}

	switch c.Style {
	case StylePixel, StyleCartoon, StyleNeon, StyleMinimal:
	default:
		return fmt.Errorf("%w: unknown style %q", ErrInvalidInput, c.Style)
	}

	return nil
}

func (c GameConfig) Apply(update GameConfigUpdate) (GameConfig, error) {
	if update.Tempo != nil {
		c.Tempo = *update.Tempo
	}

	if update.Palette != nil {
		c.Palette = *update.Palette
	}

	if update.MapSize != nil {
		c.MapSize = *update.MapSize
	}

	if update.Difficulty != nil {
		c.Difficulty = *update.Difficulty
	}

	if update.Style != nil {
		c.Style = *update.Style
	}

	if err := c.Validate(); err != nil {
		return GameConfig{}, fmt.Errorf("c.Validate: %w", err)
	}

	return c, nil
}
