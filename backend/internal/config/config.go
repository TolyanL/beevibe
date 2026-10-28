package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/TolyanL/beevibe/internal/adapter/auth"
	"github.com/TolyanL/beevibe/internal/adapter/kafka"
	"github.com/TolyanL/beevibe/internal/adapter/validator"
	"github.com/TolyanL/beevibe/pkg/httpserver"
	"github.com/TolyanL/beevibe/pkg/llm"
	"github.com/TolyanL/beevibe/pkg/logger"
	"github.com/TolyanL/beevibe/pkg/postgres"
)

// DefaultPath — путь к файлу конфигурации относительно рабочего каталога.
const DefaultPath = "configs/config.yaml"

// PathEnv — переменная окружения с путём к файлу конфигурации.
const PathEnv = "APP_CONFIG_PATH"

type App struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Tokens struct {
	GrantSignup   int64 `yaml:"grant_signup"`
	PriceGenerate int64 `yaml:"price_generate"`
	PricePatch    int64 `yaml:"price_patch"`
	PriceMeta     int64 `yaml:"price_meta"`
	TopupMin      int64 `yaml:"topup_min"`
	TopupMax      int64 `yaml:"topup_max"`
	TopupDailyCap int64 `yaml:"topup_daily_cap"`
	DemoTopup     bool  `yaml:"demo_topup"`
}

type LLMReasoning struct {
	Analyzer  llm.Effort `yaml:"analyzer"`
	Planner   llm.Effort `yaml:"planner"`
	Coder     llm.Effort `yaml:"coder"`
	Assets    llm.Effort `yaml:"assets"`
	Reviewer  llm.Effort `yaml:"reviewer"`
	Assembler llm.Effort `yaml:"assembler"`
	Patch     llm.Effort `yaml:"patch"`
	Repair    llm.Effort `yaml:"repair"`
	Meta      llm.Effort `yaml:"meta"`
}

func (c LLMReasoning) Validate() error {
	fields := []struct {
		name  string
		value llm.Effort
	}{
		{"llm_reasoning.analyzer", c.Analyzer},
		{"llm_reasoning.planner", c.Planner},
		{"llm_reasoning.coder", c.Coder},
		{"llm_reasoning.assets", c.Assets},
		{"llm_reasoning.reviewer", c.Reviewer},
		{"llm_reasoning.assembler", c.Assembler},
		{"llm_reasoning.patch", c.Patch},
		{"llm_reasoning.repair", c.Repair},
	}

	for _, field := range fields {
		if !field.value.Valid() {
			return fmt.Errorf("%s: invalid effort %q", field.name, field.value)
		}
	}

	return nil
}

type Worker struct {
	Concurrency     int           `yaml:"concurrency"`
	Timeout         time.Duration `yaml:"timeout"`
	MaxCallsPerGame int           `yaml:"max_calls_per_game"`
	MaxFixAttempts  int           `yaml:"max_fix_attempts"`
}

type Config struct {
	App          App               `yaml:"app"`
	Auth         auth.Config       `yaml:"auth"`
	Kafka        kafka.Config      `yaml:"kafka"`
	HTTP         httpserver.Config `yaml:"http"`
	Logger       logger.Config     `yaml:"logger"`
	Postgres     postgres.Config   `yaml:"postgres"`
	LLM          llm.Config        `yaml:"llm"`
	LLMReasoning LLMReasoning      `yaml:"llm_reasoning"`
	Tokens       Tokens            `yaml:"tokens"`
	Validator    validator.Config  `yaml:"validator"`
	Worker       Worker            `yaml:"worker"`
}

// New — читает конфигурацию из файла, путь берётся из APP_CONFIG_PATH.
func New() (Config, error) {
	return Load(Path())
}

// Path — путь к файлу конфигурации: APP_CONFIG_PATH или путь по умолчанию.
func Path() string {
	if path := os.Getenv(PathEnv); path != "" {
		return path
	}

	return DefaultPath
}

// Load — читает YAML-файл конфигурации и проверяет его.
func Load(path string) (Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("os.ReadFile: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("Validate: %w", err)
	}

	return cfg, nil
}

// Defaults — значения для полей, которых нет в файле конфигурации.
func Defaults() Config {
	return Config{
		App: App{
			Name:    "beevibe",
			Version: "0.1.0",
		},
		Auth: auth.Config{
			BcryptCost: 10,
		},
		Kafka: kafka.Config{
			Brokers:      []string{"localhost:9094"},
			Topic:        "game.events",
			GroupID:      "beevibe",
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
		HTTP: httpserver.Config{
			Port:            "8080",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    0,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			StaticDir:       "static",
		},
		Logger: logger.Config{
			Level:  "info",
			Format: "console",
		},
		Postgres: postgres.Config{
			DSN:            "postgres://bee:12345@localhost:5432/bee?sslmode=disable",
			MaxConns:       10,
			MinConns:       2,
			ConnectTimeout: 5 * time.Second,
		},
		LLM: llm.Config{
			Model:            "deepseek/deepseek-chat",
			BaseURL:          "https://openrouter.ai/api/v1",
			Timeout:          120 * time.Second,
			MaxRetries:       2,
			RetryBackoff:     2 * time.Second,
			ReasoningDialect: llm.ReasoningDialectOpenRouter,
		},
		Tokens: Tokens{
			GrantSignup:   1_000_000,
			PriceGenerate: 320_000,
			PricePatch:    40_000,
			PriceMeta:     10_000,
			TopupMin:      10_000,
			TopupMax:      1_000_000,
			TopupDailyCap: 5_000_000,
			DemoTopup:     true,
		},
		Worker: Worker{
			Concurrency:     2,
			Timeout:         300 * time.Second,
			MaxCallsPerGame: 12,
			MaxFixAttempts:  2,
		},
	}
}

// Validate — проверяет значения, без которых сервис не может стартовать.
func (c Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if c.App.Version == "" {
		return fmt.Errorf("app.version is required")
	}

	if c.LLM.APIKey == "" {
		return fmt.Errorf("llm.api_key is required")
	}

	if c.Postgres.DSN == "" {
		return fmt.Errorf("postgres.dsn is required")
	}

	if err := c.LLM.Validate(); err != nil {
		return fmt.Errorf("LLM.Validate: %w", err)
	}

	if err := c.LLMReasoning.Validate(); err != nil {
		return fmt.Errorf("LLMReasoning.Validate: %w", err)
	}

	return nil
}
