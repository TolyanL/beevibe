package llm

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 60 * time.Second

const (
	ReasoningDialectOpenRouter = "openrouter"
	ReasoningDialectOpenAI     = "openai"
)

type Config struct {
	APIKey           string        `envconfig:"OPENROUTER_API_KEY"  required:"true"`
	Model            string        `envconfig:"OPENROUTER_MODEL"   default:"deepseek/deepseek-chat"`
	BaseURL          string        `envconfig:"OPENROUTER_BASE_URL" default:"https://openrouter.ai/api/v1"`
	Timeout          time.Duration `envconfig:"LLM_TIMEOUT"         default:"120s"`
	MaxRetries       int           `envconfig:"LLM_MAX_RETRIES"     default:"2"`
	RetryBackoff     time.Duration `envconfig:"LLM_RETRY_BACKOFF"  default:"2s"`
	ReasoningDialect string        `envconfig:"LLM_REASONING_DIALECT" default:"openrouter"`
	DebugDir         string        `envconfig:"LLM_DEBUG_DIR"         default:""`
}

func (c Config) Validate() error {
	switch c.ReasoningDialect {
	case "", ReasoningDialectOpenRouter, ReasoningDialectOpenAI:
		return nil
	default:
		return fmt.Errorf("unknown reasoning dialect %q", c.ReasoningDialect)
	}
}

type Client struct {
	cfg        Config
	baseURL    string
	httpClient *http.Client
}

func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ForceAttemptHTTP2 = false
	transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	transport.TLSClientConfig = &tls.Config{NextProtos: []string{"http/1.1"}}

	return &Client{
		cfg:        cfg,
		baseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
		httpClient: &http.Client{Timeout: timeout, Transport: transport},
	}
}
