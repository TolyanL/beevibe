package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TolyanL/beevibe/internal/domain"
)

const okBody = `{"choices":[{"message":{"role":"assistant","content":"Игра готова"}}],` +
	`"usage":{"prompt_tokens":120,"completion_tokens":340}}`

func TestComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		maxRetries int
		handler    func(calls *atomic.Int32) http.HandlerFunc
		wantText   string
		wantUsage  Usage
		wantErr    bool
		wantCalls  int32
	}{
		{
			name: "successful response fills text and usage",
			handler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(okBody))
				}
			},
			wantText:  "Игра готова",
			wantUsage: Usage{PromptTokens: 120, CompletionTokens: 340},
			wantCalls: 1,
		},
		{
			name:       "too many requests is retried",
			maxRetries: 2,
			handler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					if calls.Add(1) == 1 {
						w.WriteHeader(http.StatusTooManyRequests)

						return
					}

					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(okBody))
				}
			},
			wantText:  "Игра готова",
			wantUsage: Usage{PromptTokens: 120, CompletionTokens: 340},
			wantCalls: 2,
		},
		{
			name:       "server error is wrapped into domain error",
			maxRetries: 1,
			handler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			wantErr:   true,
			wantCalls: 2,
		},
		{
			name: "invalid json is wrapped into domain error",
			handler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					_, _ = w.Write([]byte("not a json"))
				}
			},
			wantErr:   true,
			wantCalls: 1,
		},
		{
			name:       "client error is not retried",
			maxRetries: 2,
			handler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.WriteHeader(http.StatusBadRequest)
				}
			},
			wantErr:   true,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			calls := &atomic.Int32{}
			server := httptest.NewServer(tt.handler(calls))
			defer server.Close()

			client := New(Config{
				APIKey:       "test-key",
				Model:        "test-model",
				BaseURL:      server.URL,
				Timeout:      5 * time.Second,
				MaxRetries:   tt.maxRetries,
				RetryBackoff: time.Millisecond,
			})

			response, err := client.Complete(t.Context(), Request{
				System:      "system prompt",
				Prompt:      "user prompt",
				Temperature: 0.4,
			})

			if got := calls.Load(); got != tt.wantCalls {
				t.Errorf("calls: want %d, got %d", tt.wantCalls, got)
			}

			if tt.wantErr {
				if !errors.Is(err, domain.ErrLLMUnavailable) {
					t.Fatalf("error: want %v, got %v", domain.ErrLLMUnavailable, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if response.Text != tt.wantText {
				t.Errorf("text: want %q, got %q", tt.wantText, response.Text)
			}

			if response.Usage != tt.wantUsage {
				t.Errorf("usage: want %+v, got %+v", tt.wantUsage, response.Usage)
			}
		})
	}
}

func TestCompleteSendsRequest(t *testing.T) {
	t.Parallel()

	var (
		authorization string
		contentType   string
		body          chatRequest
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		contentType = r.Header.Get("Content-Type")

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(okBody))
	}))
	defer server.Close()

	client := New(Config{
		APIKey:  "secret-key",
		Model:   "test-model",
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})

	if _, err := client.Complete(t.Context(), Request{
		System:      "system prompt",
		Prompt:      "user prompt",
		Temperature: 0.4,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if authorization != "Bearer secret-key" {
		t.Errorf("authorization: want %q, got %q", "Bearer secret-key", authorization)
	}

	if contentType != "application/json" {
		t.Errorf("content type: want %q, got %q", "application/json", contentType)
	}

	if body.Model != "test-model" {
		t.Errorf("model: want %q, got %q", "test-model", body.Model)
	}

	if len(body.Messages) != 2 {
		t.Fatalf("messages: want 2, got %d", len(body.Messages))
	}

	if body.Messages[0].Role != "system" || body.Messages[0].Content != "system prompt" {
		t.Errorf("system message: got %+v", body.Messages[0])
	}

	if body.Messages[1].Role != "user" || body.Messages[1].Content != "user prompt" {
		t.Errorf("user message: got %+v", body.Messages[1])
	}

	if body.Temperature != 0.4 {
		t.Errorf("temperature: want 0.4, got %v", body.Temperature)
	}
}

// Тест фиксирует отказ от лимита длины ответа: поле max_tokens не должно
// попадать в запрос, иначе провайдер обрежет размышления модели.
func TestCompleteSendsNoMaxTokens(t *testing.T) {
	t.Parallel()

	var raw map[string]json.RawMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(okBody))
	}))
	defer server.Close()

	client := New(Config{
		APIKey:  "secret-key",
		Model:   "test-model",
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})

	if _, err := client.Complete(t.Context(), Request{
		System:      "system prompt",
		Prompt:      "user prompt",
		Temperature: 0.4,
		Reasoning:   EffortHigh,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := raw["max_tokens"]; ok {
		t.Errorf("max_tokens must not be sent: body = %v", raw)
	}
}

func TestCompleteCancelledContext(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(okBody))
	}))
	defer server.Close()

	client := New(Config{
		APIKey:  "test-key",
		Model:   "test-model",
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := client.Complete(ctx, Request{System: "system prompt", Prompt: "user prompt"})
	if !errors.Is(err, domain.ErrLLMUnavailable) {
		t.Fatalf("error: want %v, got %v", domain.ErrLLMUnavailable, err)
	}
}

func TestPrompt(t *testing.T) {
	t.Parallel()

	client := New(Config{})

	text, err := client.Prompt("analyzer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(text) == "" {
		t.Fatal("prompt is empty")
	}

	if _, err := client.Prompt("missing-prompt"); err == nil {
		t.Fatal("expected error for unknown prompt")
	}
}

func TestPromptVersion(t *testing.T) {
	t.Parallel()

	client := New(Config{})

	version := client.PromptVersion()
	if len(version) != 16 {
		t.Fatalf("version length: want 16, got %d", len(version))
	}

	if version != client.PromptVersion() {
		t.Fatal("version is not stable")
	}
}
