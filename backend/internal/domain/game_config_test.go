package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestConfigStyles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		style Style
		want  bool
	}{
		{name: "pixel", style: StylePixel, want: true},
		{name: "cartoon", style: StyleCartoon, want: true},
		{name: "neon", style: StyleNeon, want: true},
		{name: "minimal", style: StyleMinimal, want: true},
		{name: "custom is a request label, not a config value", style: StyleCustom},
		{name: "empty", style: ""},
		{name: "unknown", style: Style("retro")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.style.IsConfigStyle(); got != tt.want {
				t.Fatalf("Style(%q).IsConfigStyle() = %v, want %v", tt.style, got, tt.want)
			}
		})
	}

	if len(ConfigStyles()) != 4 {
		t.Fatalf("ConfigStyles() = %v, want the four CONFIG enum values", ConfigStyles())
	}
}

func TestGameConfigApply(t *testing.T) {
	t.Parallel()

	base := GameConfig{
		Tempo:      5,
		Palette:    []string{"#111111"},
		MapSize:    MapSizeMedium,
		Difficulty: "normal",
		Style:      StyleNeon,
	}

	tempo := Tempo(9)
	mapSize := MapSizeLarge

	updated, err := base.Apply(GameConfigUpdate{Tempo: &tempo, MapSize: &mapSize})
	if err != nil {
		t.Fatalf("Apply() error = %v, want nil", err)
	}

	if updated.Tempo != 9 {
		t.Fatalf("tempo = %d, want %d", updated.Tempo, 9)
	}

	if updated.MapSize != MapSizeLarge {
		t.Fatalf("map size = %s, want %s", updated.MapSize, MapSizeLarge)
	}

	if updated.Difficulty != "normal" {
		t.Fatalf("difficulty = %q, want %q", updated.Difficulty, "normal")
	}
}

func TestGameConfigApplyRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	base := GameConfig{Tempo: 5, MapSize: MapSizeMedium, Style: StyleNeon}

	invalidTempo := Tempo(42)
	if _, err := base.Apply(GameConfigUpdate{Tempo: &invalidTempo}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Apply() error = %v, want ErrInvalidInput", err)
	}

	invalidSize := MapSize("huge")
	if _, err := base.Apply(GameConfigUpdate{MapSize: &invalidSize}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Apply() error = %v, want ErrInvalidInput", err)
	}

	invalidStyle := Style("retro")
	if _, err := base.Apply(GameConfigUpdate{Style: &invalidStyle}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Apply() error = %v, want ErrInvalidInput", err)
	}
}

func TestClampPlaytime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seconds int
		want    int
	}{
		{name: "normal", seconds: 120, want: 120},
		{name: "over limit", seconds: 5000, want: MaxPlaytimeSeconds},
		{name: "negative", seconds: -10, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ClampPlaytime(tt.seconds); got != tt.want {
				t.Fatalf("ClampPlaytime(%d) = %d, want %d", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestNewRating(t *testing.T) {
	t.Parallel()

	if _, err := NewRating(uuid.New(), uuid.New(), RatingValue(0)); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewRating() error = %v, want ErrInvalidInput", err)
	}
}
