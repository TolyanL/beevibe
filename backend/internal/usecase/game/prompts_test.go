package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestCanonicalConfigBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  domain.GameParams
		want    []string
		notWant []string
	}{
		{
			name:   "full params carry canonical values with enums",
			params: domain.GameParams{Tempo: 5, MapSize: domain.MapSizeSmall, Style: domain.StylePixel},
			want: []string{
				"- tempo: 5",
				`- mapSize: "small" (allowed: "small", "medium", "large")`,
				`- style: "pixel" (allowed: "pixel", "cartoon", "neon", "minimal")`,
			},
		},
		{
			name:   "empty params produce no block",
			params: domain.GameParams{},
		},
		{
			name:    "tempo out of the valid range is skipped",
			params:  domain.GameParams{Tempo: 0, Style: domain.StyleNeon},
			want:    []string{`- style: "neon"`},
			notWant: []string{"- tempo:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := canonicalConfigBlock(tt.params)

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Fatalf("canonicalConfigBlock(%+v) = %q, want substring %q", tt.params, got, want)
				}
			}

			for _, notWant := range tt.notWant {
				if strings.Contains(got, notWant) {
					t.Fatalf("canonicalConfigBlock(%+v) = %q, must not contain %q", tt.params, got, notWant)
				}
			}
		})
	}
}

func TestCodeAndAssemblePromptsCarryCanonicalConfig(t *testing.T) {
	t.Parallel()

	p := &pipeline{
		game: domain.Game{
			ID:     uuid.New(),
			Prompt: "змейка",
			Params: domain.GameParams{Tempo: 5, MapSize: domain.MapSizeSmall, Style: domain.StylePixel},
		},
	}

	for name, prompt := range map[string]string{
		"code":     buildCodePrompt(p),
		"assemble": buildAssemblePrompt(p),
	} {
		if !strings.Contains(prompt, "Canonical CONFIG values") || !strings.Contains(prompt, `- style: "pixel"`) {
			t.Errorf("%s prompt must carry the canonical CONFIG values, got:\n%s", name, prompt)
		}
	}
}

func TestCodeAndAssemblePromptsOmitCanonicalConfigWhenEmpty(t *testing.T) {
	t.Parallel()

	p := &pipeline{game: domain.Game{ID: uuid.New(), Prompt: "змейка"}}

	for name, prompt := range map[string]string{
		"code":     buildCodePrompt(p),
		"assemble": buildAssemblePrompt(p),
	} {
		if strings.Contains(prompt, "Canonical CONFIG values") {
			t.Errorf("%s prompt must not invent canonical CONFIG values for empty params", name)
		}
	}
}
