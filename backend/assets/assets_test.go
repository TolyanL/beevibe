package assets

import (
	"slices"
	"strings"
	"testing"
)

func TestPromptNames(t *testing.T) {
	t.Parallel()

	want := []string{"analyzer", "assembler", "assets", "coder", "patch", "planner", "repair", "reviewer"}

	got := PromptNames()
	if !slices.Equal(got, want) {
		t.Fatalf("names: want %v, got %v", want, got)
	}
}

func TestPrompt(t *testing.T) {
	t.Parallel()

	for _, name := range PromptNames() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			text, err := Prompt(name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(text) == "" {
				t.Fatal("prompt is empty")
			}
		})
	}
}

func TestPromptUnknown(t *testing.T) {
	t.Parallel()

	if _, err := Prompt("unknown-prompt"); err == nil {
		t.Fatal("expected error for unknown prompt")
	}
}

func TestPromptVersion(t *testing.T) {
	t.Parallel()

	version := PromptVersion()
	if len(version) != versionLength {
		t.Fatalf("version length: want %d, got %d", versionLength, len(version))
	}

	if version != PromptVersion() {
		t.Fatal("version is not stable")
	}
}
