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
	APIKey           string        `yaml:"api_key"`
	Model            string        `yaml:"model"`
	BaseURL          string        `yaml:"base_url"`
	Timeout          time.Duration `yaml:"timeout"`
	MaxRetries       int           `yaml:"max_retries"`
	RetryBackoff     time.Duration `yaml:"retry_backoff"`
	ReasoningDialect string        `yaml:"reasoning_dialect"`
	DebugDir         string        `yaml:"debug_dir"`
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
