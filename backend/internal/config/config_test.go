package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TolyanL/beevibe/pkg/llm"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	return path
}

const minimalConfig = `
app:
  name: test
  version: test
llm:
  api_key: test-key
`

func TestLoadReadsValues(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, `
app:
  name: gamegen
  version: 0.2.0
http:
  port: "8099"
  write_timeout: 0s
  idle_timeout: 90s
postgres:
  dsn: postgres://bee@localhost:5432/bee
  max_conns: 20
llm:
  api_key: secret
  model: qwen3.8-27b
  base_url: http://localhost:12286/v1
  timeout: 1800s
  max_retries: 6
  reasoning_dialect: openai
  debug_dir: /tmp/dump
llm_reasoning:
  analyzer: low
  coder: high
tokens:
  price_generate: 1000
worker:
  concurrency: 4
  timeout: 7200s
  max_calls_per_game: 16
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.App.Name != "gamegen" || cfg.App.Version != "0.2.0" {
		t.Errorf("app = %+v, want gamegen 0.2.0", cfg.App)
	}

	if cfg.HTTP.Port != "8099" || cfg.HTTP.IdleTimeout != 90*time.Second {
		t.Errorf("http = %+v, want port 8099 idle 90s", cfg.HTTP)
	}

	if cfg.Postgres.MaxConns != 20 {
		t.Errorf("max conns = %d, want 20", cfg.Postgres.MaxConns)
	}

	if cfg.LLM.APIKey != "secret" || cfg.LLM.Model != "qwen3.8-27b" || cfg.LLM.Timeout != 1800*time.Second {
		t.Errorf("llm = %+v, want secret qwen3.8-27b 1800s", cfg.LLM)
	}

	if cfg.LLM.ReasoningDialect != llm.ReasoningDialectOpenAI {
		t.Errorf("dialect = %q, want %q", cfg.LLM.ReasoningDialect, llm.ReasoningDialectOpenAI)
	}

	if cfg.LLM.DebugDir != "/tmp/dump" {
		t.Errorf("debug dir = %q, want /tmp/dump", cfg.LLM.DebugDir)
	}

	if cfg.LLMReasoning.Analyzer != llm.EffortLow || cfg.LLMReasoning.Coder != llm.EffortHigh {
		t.Errorf("reasoning = %+v, want analyzer low coder high", cfg.LLMReasoning)
	}

	if cfg.Tokens.PriceGenerate != 1000 {
		t.Errorf("price generate = %d, want 1000", cfg.Tokens.PriceGenerate)
	}

	if cfg.Worker.Concurrency != 4 || cfg.Worker.Timeout != 7200*time.Second || cfg.Worker.MaxCallsPerGame != 16 {
		t.Errorf("worker = %+v, want 4 7200s 16", cfg.Worker)
	}
}

func TestLoadKeepsDefaultsForAbsentFields(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, minimalConfig)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.HTTP.Port != Defaults().HTTP.Port {
		t.Errorf("http port = %q, want %q", cfg.HTTP.Port, Defaults().HTTP.Port)
	}

	if cfg.Worker.MaxFixAttempts != Defaults().Worker.MaxFixAttempts {
		t.Errorf("max fix attempts = %d, want %d", cfg.Worker.MaxFixAttempts, Defaults().Worker.MaxFixAttempts)
	}

	if cfg.LLMReasoning.Analyzer != llm.EffortDefault {
		t.Errorf("analyzer = %q, want default", cfg.LLMReasoning.Analyzer)
	}
}

func TestLoadFailsOnMissingFile(t *testing.T) {
	t.Parallel()

	_, err := Load(filepath.Join(t.TempDir(), "absent.yaml"))
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadFailsOnBrokenYaml(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, "app: [unclosed\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadFailsOnInvalidEffort(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, `
app:
  name: test
  version: test
llm:
  api_key: test-key
llm_reasoning:
  coder: turbo
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "llm_reasoning.coder") {
		t.Fatalf("error = %v, want field name", err)
	}
}

func TestLoadFailsOnInvalidDialect(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, `
app:
  name: test
  version: test
llm:
  api_key: test-key
  reasoning_dialect: bogus
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "reasoning dialect") {
		t.Fatalf("error = %v, want dialect", err)
	}
}

func TestLoadFailsWithoutApiKey(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, `
app:
  name: test
  version: test
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "llm.api_key") {
		t.Fatalf("error = %v, want api key", err)
	}
}

func TestPathUsesEnvironment(t *testing.T) {
	t.Setenv(PathEnv, "/tmp/custom.yaml")

	if got := Path(); got != "/tmp/custom.yaml" {
		t.Errorf("Path() = %q, want /tmp/custom.yaml", got)
	}
}

func TestPathFallsBackToDefault(t *testing.T) {
	t.Setenv(PathEnv, "")

	if got := Path(); got != DefaultPath {
		t.Errorf("Path() = %q, want %q", got, DefaultPath)
	}
}
