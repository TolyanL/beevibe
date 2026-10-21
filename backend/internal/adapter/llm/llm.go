package llm

import (
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 60 * time.Second

// Config — параметры адаптера чат-комплишена.
type Config struct {
	APIKey       string        `envconfig:"OPENROUTER_API_KEY"  required:"true"`
	Model        string        `envconfig:"OPENROUTER_MODEL"   default:"deepseek/deepseek-chat"`
	BaseURL      string        `envconfig:"OPENROUTER_BASE_URL" default:"https://openrouter.ai/api/v1"`
	Timeout      time.Duration `envconfig:"LLM_TIMEOUT"         default:"120s"`
	MaxRetries   int           `envconfig:"LLM_MAX_RETRIES"     default:"2"`
	RetryBackoff time.Duration `envconfig:"LLM_RETRY_BACKOFF"  default:"2s"`
}

// Client — реализация порта usecase.LLM поверх OpenRouter-совместимого API.
type Client struct {
	cfg        Config
	baseURL    string
	httpClient *http.Client
}

// New — конструктор адаптера: собирает HTTP-клиент с таймаутом из конфига.
func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	return &Client{
		cfg:        cfg,
		baseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}
