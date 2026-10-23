package validator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestValidatorApplyConfig(t *testing.T) {
	t.Parallel()

	v := New(Config{})

	base, err := v.ExtractConfig(context.Background(), validGameHTML)
	if err != nil {
		t.Fatalf("ExtractConfig() error = %v", err)
	}

	fast := base
	fast.Tempo = 8

	tests := []struct {
		name      string
		html      string
		cfg       domain.GameConfig
		wantErr   bool
		wantTempo domain.Tempo
		wantParts []string
	}{
		{
			name:      "tempo is replaced",
			html:      validGameHTML,
			cfg:       fast,
			wantTempo: 8,
		},
		{
			name:    "config block is missing",
			html:    gameWithoutConfig(),
			cfg:     fast,
			wantErr: true,
		},
		{
			name:      "missing key is added",
			html:      strings.Replace(validGameHTML, "      difficulty: \"normal\",\n", "", 1),
			cfg:       base,
			wantTempo: 5,
			wantParts: []string{"difficulty: \"normal\""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			updated, err := v.ApplyConfig(context.Background(), tt.html, tt.cfg)
			if tt.wantErr {
				if !errors.Is(err, domain.ErrValidationFailed) {
					t.Fatalf("ApplyConfig() error = %v, want %v", err, domain.ErrValidationFailed)
				}

				return
			}

			if err != nil {
				t.Fatalf("ApplyConfig() error = %v", err)
			}

			if updated == tt.html {
				t.Fatal("ApplyConfig() did not change the document")
			}

			for _, part := range tt.wantParts {
				if !strings.Contains(updated, part) {
					t.Errorf("ApplyConfig() result does not contain %q", part)
				}
			}

			applied, err := v.ExtractConfig(context.Background(), updated)
			if err != nil {
				t.Fatalf("ExtractConfig() error = %v", err)
			}

			if applied.Tempo != tt.wantTempo {
				t.Errorf("ExtractConfig() tempo = %d, want %d", applied.Tempo, tt.wantTempo)
			}

			if applied.Palette == nil || len(applied.Palette) != len(tt.cfg.Palette) {
				t.Errorf("ExtractConfig() palette = %v, want %v", applied.Palette, tt.cfg.Palette)
			}

			if applied.MapSize != tt.cfg.MapSize {
				t.Errorf("ExtractConfig() map size = %q, want %q", applied.MapSize, tt.cfg.MapSize)
			}

			if applied.Difficulty != tt.cfg.Difficulty {
				t.Errorf("ExtractConfig() difficulty = %q, want %q", applied.Difficulty, tt.cfg.Difficulty)
			}

			if applied.Controls != tt.cfg.Controls {
				t.Errorf("ExtractConfig() controls = %q, want %q", applied.Controls, tt.cfg.Controls)
			}

			if applied.Style != tt.cfg.Style {
				t.Errorf("ExtractConfig() style = %q, want %q", applied.Style, tt.cfg.Style)
			}
		})
	}
}

func TestValidatorApplyConfigSingleLineBlock(t *testing.T) {
	t.Parallel()

	const html = `<script>const CONFIG = { tempo: 5, palette: ["#fff"], mapSize: "small", difficulty: "easy", controls: "touch", style: "pixel" };</script>`

	v := New(Config{})

	cfg, err := v.ExtractConfig(context.Background(), html)
	if err != nil {
		t.Fatalf("ExtractConfig() error = %v", err)
	}

	cfg.Tempo = 8

	updated, err := v.ApplyConfig(context.Background(), html, cfg)
	if err != nil {
		t.Fatalf("ApplyConfig() error = %v", err)
	}

	if !strings.Contains(updated, "tempo: 8") {
		t.Errorf("ApplyConfig() result = %q", updated)
	}

	applied, err := v.ExtractConfig(context.Background(), updated)
	if err != nil {
		t.Fatalf("ExtractConfig() error = %v", err)
	}

	if applied.Tempo != 8 {
		t.Errorf("ExtractConfig() tempo = %d, want 8", applied.Tempo)
	}
}

func TestValidatorApplyConfigAddsKeyToSingleLineBlock(t *testing.T) {
	t.Parallel()

	const html = `<script>const CONFIG = { tempo: 5, mapSize: "small", style: "pixel" };</script>`

	v := New(Config{})

	updated, err := v.ApplyConfig(context.Background(), html, domain.GameConfig{
		Tempo:      6,
		Palette:    []string{"#fff"},
		MapSize:    domain.MapSizeSmall,
		Difficulty: "easy",
		Controls:   "touch",
		Style:      domain.StylePixel,
	})
	if err != nil {
		t.Fatalf("ApplyConfig() error = %v", err)
	}

	if !strings.Contains(updated, `difficulty: "easy"`) {
		t.Errorf("ApplyConfig() result = %q", updated)
	}

	applied, err := v.ExtractConfig(context.Background(), updated)
	if err != nil {
		t.Fatalf("ExtractConfig() error = %v", err)
	}

	if applied.Difficulty != "easy" {
		t.Errorf("ExtractConfig() difficulty = %q, want %q", applied.Difficulty, "easy")
	}
}

func TestValidatorApplyConfigKeepsFormatting(t *testing.T) {
	t.Parallel()

	v := New(Config{})

	cfg, err := v.ExtractConfig(context.Background(), validGameHTML)
	if err != nil {
		t.Fatalf("ExtractConfig() error = %v", err)
	}

	cfg.Tempo = 9

	updated, err := v.ApplyConfig(context.Background(), validGameHTML, cfg)
	if err != nil {
		t.Fatalf("ApplyConfig() error = %v", err)
	}

	if !strings.Contains(updated, "      tempo: 9,\n") {
		t.Errorf("ApplyConfig() result keeps no formatting:\n%s", updated)
	}

	if !strings.Contains(updated, "      style: \"neon\"\n") {
		t.Errorf("ApplyConfig() result keeps no formatting:\n%s", updated)
	}
}
