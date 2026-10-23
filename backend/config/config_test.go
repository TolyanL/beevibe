package config

import (
	"strings"
	"testing"

	"github.com/TolyanL/beevibe/pkg/llm"
)

func TestNewReasoningConfig(t *testing.T) {
	t.Setenv("APP_NAME", "test")
	t.Setenv("APP_VERSION", "test")
	t.Setenv("OPENROUTER_API_KEY", "test-key")
	t.Setenv("LLM_REASONING_DIALECT", llm.ReasoningDialectOpenRouter)

	for _, name := range []string{
		"LLM_REASONING_EFFORT_ANALYZER",
		"LLM_REASONING_EFFORT_PLANNER",
		"LLM_REASONING_EFFORT_CODER",
		"LLM_REASONING_EFFORT_ASSETS",
		"LLM_REASONING_EFFORT_REVIEWER",
		"LLM_REASONING_EFFORT_ASSEMBLER",
		"LLM_REASONING_EFFORT_PATCH",
		"LLM_REASONING_EFFORT_REPAIR",
	} {
		t.Setenv(name, "default")
	}

	t.Run("valid efforts", func(t *testing.T) {
		t.Setenv("LLM_REASONING_EFFORT_ANALYZER", "low")
		t.Setenv("LLM_REASONING_EFFORT_CODER", "xhigh")

		cfg, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if cfg.LLMReasoning.Analyzer != llm.EffortLow {
			t.Errorf("analyzer = %q, want %q", cfg.LLMReasoning.Analyzer, llm.EffortLow)
		}

		if cfg.LLMReasoning.Coder != llm.EffortXHigh {
			t.Errorf("coder = %q, want %q", cfg.LLMReasoning.Coder, llm.EffortXHigh)
		}

		if cfg.LLMReasoning.Planner != llm.Effort("default") {
			t.Errorf("planner = %q, want default", cfg.LLMReasoning.Planner)
		}

		if cfg.LLM.ReasoningDialect != llm.ReasoningDialectOpenRouter {
			t.Errorf("dialect = %q, want %q", cfg.LLM.ReasoningDialect, llm.ReasoningDialectOpenRouter)
		}
	})

	t.Run("invalid effort fails fast", func(t *testing.T) {
		t.Setenv("LLM_REASONING_EFFORT_CODER", "turbo")

		_, err := New()
		if err == nil {
			t.Fatal("New() error = nil, want error")
		}

		if !strings.Contains(err.Error(), "LLM_REASONING_EFFORT_CODER") {
			t.Fatalf("error = %v, want field name", err)
		}
	})

	t.Run("invalid dialect fails fast", func(t *testing.T) {
		t.Setenv("LLM_REASONING_DIALECT", "bogus")

		_, err := New()
		if err == nil {
			t.Fatal("New() error = nil, want error")
		}

		if !strings.Contains(err.Error(), "reasoning dialect") {
			t.Fatalf("error = %v, want dialect", err)
		}
	})
}
