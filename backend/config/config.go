package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/TolyanL/beevibe/internal/adapter/auth"
	"github.com/TolyanL/beevibe/internal/adapter/kafka"
	"github.com/TolyanL/beevibe/internal/adapter/validator"
	"github.com/TolyanL/beevibe/pkg/httpserver"
	"github.com/TolyanL/beevibe/pkg/llm"
	"github.com/TolyanL/beevibe/pkg/logger"
	"github.com/TolyanL/beevibe/pkg/postgres"
)

type App struct {
	Name    string `envconfig:"APP_NAME"    required:"true"`
	Version string `envconfig:"APP_VERSION" required:"true"`
}

type Tokens struct {
	GrantSignup   int64 `envconfig:"TOKENS_GRANT_SIGNUP"   default:"1000000"`
	GrantTopup    int64 `envconfig:"TOKENS_GRANT_TOPUP"    default:"500000"`
	PriceGenerate int64 `envconfig:"TOKENS_PRICE_GENERATE" default:"320000"`
	PricePatch    int64 `envconfig:"TOKENS_PRICE_PATCH"    default:"40000"`
}

type LLMReasoning struct {
	Analyzer  llm.Effort `envconfig:"LLM_REASONING_EFFORT_ANALYZER"  default:"default"`
	Planner   llm.Effort `envconfig:"LLM_REASONING_EFFORT_PLANNER"   default:"default"`
	Coder     llm.Effort `envconfig:"LLM_REASONING_EFFORT_CODER"     default:"default"`
	Assets    llm.Effort `envconfig:"LLM_REASONING_EFFORT_ASSETS"    default:"default"`
	Reviewer  llm.Effort `envconfig:"LLM_REASONING_EFFORT_REVIEWER"  default:"default"`
	Assembler llm.Effort `envconfig:"LLM_REASONING_EFFORT_ASSEMBLER" default:"default"`
	Patch     llm.Effort `envconfig:"LLM_REASONING_EFFORT_PATCH"     default:"default"`
	Repair    llm.Effort `envconfig:"LLM_REASONING_EFFORT_REPAIR"    default:"default"`
}

func (c LLMReasoning) Validate() error {
	fields := []struct {
		name  string
		value llm.Effort
	}{
		{"LLM_REASONING_EFFORT_ANALYZER", c.Analyzer},
		{"LLM_REASONING_EFFORT_PLANNER", c.Planner},
		{"LLM_REASONING_EFFORT_CODER", c.Coder},
		{"LLM_REASONING_EFFORT_ASSETS", c.Assets},
		{"LLM_REASONING_EFFORT_REVIEWER", c.Reviewer},
		{"LLM_REASONING_EFFORT_ASSEMBLER", c.Assembler},
		{"LLM_REASONING_EFFORT_PATCH", c.Patch},
		{"LLM_REASONING_EFFORT_REPAIR", c.Repair},
	}

	for _, field := range fields {
		if !field.value.Valid() {
			return fmt.Errorf("%s: invalid effort %q", field.name, field.value)
		}
	}

	return nil
}

type Worker struct {
	Concurrency     int           `envconfig:"GENERATION_CONCURRENCY"      default:"2"`
	Timeout         time.Duration `envconfig:"GENERATION_TIMEOUT"          default:"300s"`
	MaxCallsPerGame int           `envconfig:"LLM_MAX_CALLS_PER_GAME"      default:"12"`
	MaxFixAttempts  int           `envconfig:"GENERATION_MAX_FIX_ATTEMPTS" default:"2"`
}

type Config struct {
	App          App
	Auth         auth.Config
	Kafka        kafka.Config
	HTTP         httpserver.Config
	Logger       logger.Config
	Postgres     postgres.Config
	LLM          llm.Config
	LLMReasoning LLMReasoning
	Tokens       Tokens
	Validator    validator.Config
	Worker       Worker
}

func New() (Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("envconfig.Process: %w", err)
	}

	if err := cfg.LLM.Validate(); err != nil {
		return Config{}, fmt.Errorf("LLM.Validate: %w", err)
	}

	if err := cfg.LLMReasoning.Validate(); err != nil {
		return Config{}, fmt.Errorf("LLMReasoning.Validate: %w", err)
	}

	return cfg, nil
}

func NewMigrate() (postgres.Config, error) {
	var cfg postgres.Config

	if err := envconfig.Process("", &cfg); err != nil {
		return postgres.Config{}, fmt.Errorf("envconfig.Process: %w", err)
	}

	return cfg, nil
}
