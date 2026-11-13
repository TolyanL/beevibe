package domain

import (
	"errors"
	"testing"
)

func TestNewSpeedDefaultsToNormal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  Speed
	}{
		{name: "empty means normal", value: "", want: SpeedNormal},
		{name: "normal", value: "normal", want: SpeedNormal},
		{name: "fast", value: "fast", want: SpeedFast},
		{name: "upper case", value: " FAST ", want: SpeedFast},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSpeed(tt.value)
			if err != nil {
				t.Fatalf("NewSpeed() error = %v, want nil", err)
			}

			if got != tt.want {
				t.Fatalf("NewSpeed() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewSpeedRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	_, err := NewSpeed("turbo")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewSpeed() error = %v, want ErrInvalidInput", err)
	}
}

func TestGameParamsValidateRejectsUnknownSpeed(t *testing.T) {
	t.Parallel()

	params := GameParams{Speed: Speed("turbo")}

	if err := params.Validate(); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Validate() error = %v, want ErrInvalidInput", err)
	}
}

func TestGameParamsValidateAcceptsEmptySpeed(t *testing.T) {
	t.Parallel()

	params := GameParams{Genre: GenreSnake}

	if err := params.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
