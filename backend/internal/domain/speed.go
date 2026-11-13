package domain

import (
	"fmt"
	"strings"
)

type Speed string

const (
	SpeedNormal Speed = "normal"
	SpeedFast   Speed = "fast"
)

func NewSpeed(value string) (Speed, error) {
	normalized := Speed(strings.ToLower(strings.TrimSpace(value)))

	if normalized == "" {
		return SpeedNormal, nil
	}

	if !normalized.Valid() {
		return "", fmt.Errorf("%w: unknown speed %q", ErrInvalidInput, value)
	}

	return normalized, nil
}

func (s Speed) Valid() bool {
	switch s {
	case "", SpeedNormal, SpeedFast:
		return true
	default:
		return false
	}
}

func (s Speed) IsFast() bool {
	return s == SpeedFast
}
