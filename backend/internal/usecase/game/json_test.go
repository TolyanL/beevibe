package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestParseSpecPlanSplitsFenceAndPlan(t *testing.T) {
	t.Parallel()

	spec, plan, err := parseSpecPlan("```json\n" + analyzerJSON + "\n```\n\n## План\n\n- state: menu -> playing")
	if err != nil {
		t.Fatalf("parseSpecPlan() error = %v, want nil", err)
	}

	if spec.Genre != "snake" {
		t.Fatalf("genre = %q, want snake", spec.Genre)
	}

	if !strings.HasPrefix(plan, "## План") {
		t.Fatalf("plan = %q, want the markdown after the fence", plan)
	}
}

func TestParseSpecPlanKeepsBracesFromThePlan(t *testing.T) {
	t.Parallel()

	answer := analyzerJSON + "\n\n## План\n\n`const CONFIG = { tempo: 5 };`"

	spec, plan, err := parseSpecPlan(answer)
	if err != nil {
		t.Fatalf("parseSpecPlan() error = %v, want nil", err)
	}

	if spec.Genre != "snake" {
		t.Fatalf("genre = %q, want snake", spec.Genre)
	}

	if !strings.Contains(plan, "tempo: 5") {
		t.Fatalf("plan = %q, want the braces of the plan preserved", plan)
	}
}

func TestParseSpecPlanRejectsEmptyAnswer(t *testing.T) {
	t.Parallel()

	_, _, err := parseSpecPlan("   ")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("parseSpecPlan() error = %v, want ErrInvalidInput", err)
	}
}

func TestParseSpecPlanRejectsMissingObject(t *testing.T) {
	t.Parallel()

	_, _, err := parseSpecPlan("no json here")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("parseSpecPlan() error = %v, want ErrInvalidInput", err)
	}
}

func TestParseSpecPlanRejectsBrokenJSON(t *testing.T) {
	t.Parallel()

	_, _, err := parseSpecPlan("{not json}")
	if err == nil {
		t.Fatal("parseSpecPlan() error = nil, want error")
	}
}

func TestBuildOneShotPromptCarriesSpecPlanAndConfig(t *testing.T) {
	t.Parallel()

	p := &pipeline{
		spec: domain.GameSpec{Genre: "snake", Mechanic: "eat and grow"},
		plan: "## План",
		game: domain.Game{Params: domain.GameParams{Tempo: 7, MapSize: domain.MapSizeSmall, Style: domain.StyleNeon}},
	}

	prompt := buildOneShotPrompt(p)

	for _, want := range []string{`"genre":"snake"`, "## План", "tempo: 7", "inline SVG"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt = %q, want %q", prompt, want)
		}
	}
}

func TestBuildSpecPlanPromptAsksForOneAnswer(t *testing.T) {
	t.Parallel()

	prompt := buildSpecPlanPrompt(domain.Game{Prompt: "неоновая змейка"})

	if !strings.Contains(prompt, "неоновая змейка") {
		t.Fatalf("prompt = %q, want the user request inside", prompt)
	}

	if !strings.Contains(prompt, "implementation plan") {
		t.Fatalf("prompt = %q, want the plan instruction", prompt)
	}
}
