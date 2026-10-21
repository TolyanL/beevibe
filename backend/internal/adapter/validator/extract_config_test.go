package validator

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestValidatorExtractConfig(t *testing.T) {
	t.Parallel()

	valid := domain.GameConfig{
		Tempo:      5,
		Palette:    []string{"#101018", "#3a3a55", "#39ff14"},
		MapSize:    domain.MapSizeMedium,
		Difficulty: "normal",
		Controls:   "touch+keyboard",
		Style:      domain.StyleNeon,
	}

	tests := []struct {
		name    string
		html    string
		want    domain.GameConfig
		wantErr bool
	}{
		{
			name: "valid game",
			html: validGameHTML,
			want: valid,
		},
		{
			name: "snake case keys",
			html: strings.Replace(
				strings.Replace(validGameHTML, `mapSize: "medium"`, `map_size: "medium"`, 1),
				`controls: "touch+keyboard"`, `control: "touch+keyboard"`, 1,
			),
			want: valid,
		},
		{
			name:    "no config block",
			html:    gameWithoutConfig(),
			wantErr: true,
		},
		{
			name:    "unknown map size",
			html:    strings.Replace(validGameHTML, `mapSize: "medium"`, `mapSize: "gigantic"`, 1),
			wantErr: true,
		},
		{
			name:    "unknown style",
			html:    strings.Replace(validGameHTML, `style: "neon"`, `style: "chalk"`, 1),
			wantErr: true,
		},
		{
			name:    "tempo out of range",
			html:    strings.Replace(validGameHTML, "tempo: 5", "tempo: 42", 1),
			wantErr: true,
		},
		{
			name:    "tempo is not a number",
			html:    strings.Replace(validGameHTML, "tempo: 5", `tempo: "fast"`, 1),
			wantErr: true,
		},
		{
			name:    "palette is not an array",
			html:    strings.Replace(validGameHTML, `palette: ["#101018", "#3a3a55", "#39ff14"]`, `palette: "#101018"`, 1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := New(Config{})

			cfg, err := v.ExtractConfig(context.Background(), tt.html)
			if tt.wantErr {
				if !errors.Is(err, domain.ErrValidationFailed) {
					t.Fatalf("ExtractConfig() error = %v, want %v", err, domain.ErrValidationFailed)
				}

				return
			}

			if err != nil {
				t.Fatalf("ExtractConfig() error = %v", err)
			}

			if cfg.Tempo != tt.want.Tempo {
				t.Errorf("ExtractConfig() tempo = %d, want %d", cfg.Tempo, tt.want.Tempo)
			}

			if cfg.MapSize != tt.want.MapSize {
				t.Errorf("ExtractConfig() map size = %q, want %q", cfg.MapSize, tt.want.MapSize)
			}

			if cfg.Style != tt.want.Style {
				t.Errorf("ExtractConfig() style = %q, want %q", cfg.Style, tt.want.Style)
			}

			if cfg.Difficulty != tt.want.Difficulty {
				t.Errorf("ExtractConfig() difficulty = %q, want %q", cfg.Difficulty, tt.want.Difficulty)
			}

			if cfg.Controls != tt.want.Controls {
				t.Errorf("ExtractConfig() controls = %q, want %q", cfg.Controls, tt.want.Controls)
			}

			if !slices.Equal(cfg.Palette, tt.want.Palette) {
				t.Errorf("ExtractConfig() palette = %v, want %v", cfg.Palette, tt.want.Palette)
			}
		})
	}
}
