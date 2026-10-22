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

type LLMTokens struct {
	Analyzer  int `envconfig:"LLM_MAX_TOKENS_ANALYZER"  default:"4000"`
	Planner   int `envconfig:"LLM_MAX_TOKENS_PLANNER"   default:"4000"`
	Coder     int `envconfig:"LLM_MAX_TOKENS_CODER"     default:"16000"`
	Assets    int `envconfig:"LLM_MAX_TOKENS_ASSETS"    default:"8000"`
	Reviewer  int `envconfig:"LLM_MAX_TOKENS_REVIEWER"  default:"4000"`
	Assembler int `envconfig:"LLM_MAX_TOKENS_ASSEMBLER" default:"8000"`
	Patch     int `envconfig:"LLM_MAX_TOKENS_PATCH"     default:"16000"`
	Repair    int `envconfig:"LLM_MAX_TOKENS_REPAIR"    default:"16000"`
}

type Worker struct {
	Concurrency     int           `envconfig:"GENERATION_CONCURRENCY"      default:"2"`
	Timeout         time.Duration `envconfig:"GENERATION_TIMEOUT"          default:"300s"`
	MaxCallsPerGame int           `envconfig:"LLM_MAX_CALLS_PER_GAME"      default:"12"`
	MaxFixAttempts  int           `envconfig:"GENERATION_MAX_FIX_ATTEMPTS" default:"2"`
}

type Config struct {
	App       App
	Auth      auth.Config
	Kafka     kafka.Config
	HTTP      httpserver.Config
	Logger    logger.Config
	Postgres  postgres.Config
	LLM       llm.Config
	LLMToken  LLMTokens
	Tokens    Tokens
	Validator validator.Config
	Worker    Worker
}

func New() (Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("envconfig.Process: %w", err)
	}

	return cfg, nil
}

// NewMigrate — конфиг CLI миграций: нужен только доступ к БД.
func NewMigrate() (postgres.Config, error) {
	var cfg postgres.Config

	if err := envconfig.Process("", &cfg); err != nil {
		return postgres.Config{}, fmt.Errorf("envconfig.Process: %w", err)
	}

	return cfg, nil
}
