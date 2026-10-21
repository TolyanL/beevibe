package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

const (
	chatCompletionsPath = "/chat/completions"
	defaultMaxTokens    = 4096
	maxErrorBodySize    = 2048
	maxRetryDelay       = 30 * time.Second
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
}

// Complete — чат-комплишен с ретраями по 429/5xx и сетевым ошибкам.
func (c *Client) Complete(ctx context.Context, req dto.LLMRequest) (dto.LLMResponse, error) {
	ctx, span := tracer.Start(ctx, "adapter llm Complete")
	defer span.End()

	var response dto.LLMResponse

	body, err := c.chatBody(req)
	if err != nil {
		span.SetError(err)

		return response, fmt.Errorf("chatBody: %w", err)
	}

	var lastErr error

	for attempt := range max(c.cfg.MaxRetries, 0) + 1 {
		if attempt > 0 {
			if err := c.wait(ctx, attempt); err != nil {
				span.SetError(err)

				return response, fmt.Errorf("wait: %w", unavailable(err))
			}
		}

		raw, retryable, err := c.send(ctx, body)
		if err == nil {
			decoded, err := c.decode(raw)
			if err != nil {
				span.SetError(err)

				return response, fmt.Errorf("decode: %w", unavailable(err))
			}

			return decoded, nil
		}

		lastErr = err

		if !retryable || ctx.Err() != nil {
			break
		}
	}

	span.SetError(lastErr)

	return response, fmt.Errorf("send: %w", unavailable(lastErr))
}

func (c *Client) chatBody(req dto.LLMRequest) ([]byte, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	body, err := json.Marshal(chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: req.System},
			{Role: "user", Content: req.Prompt},
		},
		Temperature: req.Temperature,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	return body, nil
}

func (c *Client) send(ctx context.Context, body []byte) ([]byte, bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+chatCompletionsPath, bytes.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, true, fmt.Errorf("httpClient.Do: %w", err)
	}

	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		raw, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBodySize))

		return nil, retryableStatus(response.StatusCode),
			fmt.Errorf("unexpected status %d: %s", response.StatusCode, strings.TrimSpace(string(raw)))
	}

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, true, fmt.Errorf("io.ReadAll: %w", err)
	}

	return raw, false, nil
}

func (c *Client) decode(raw []byte) (dto.LLMResponse, error) {
	var response dto.LLMResponse

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return response, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return response, fmt.Errorf("response contains no choices")
	}

	return dto.LLMResponse{
		Text: parsed.Choices[0].Message.Content,
		Usage: dto.LLMUsage{
			PromptTokens:     parsed.Usage.PromptTokens,
			CompletionTokens: parsed.Usage.CompletionTokens,
		},
	}, nil
}

func (c *Client) wait(ctx context.Context, attempt int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("ctx.Err: %w", err)
	}

	delay := c.retryDelay(attempt)
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("ctx.Done: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

func (c *Client) retryDelay(attempt int) time.Duration {
	delay := c.cfg.RetryBackoff
	if delay <= 0 {
		return 0
	}

	for range attempt - 1 {
		delay *= 2
		if delay >= maxRetryDelay {
			return maxRetryDelay
		}
	}

	return delay
}

func retryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func unavailable(err error) error {
	return fmt.Errorf("%w: %w", domain.ErrLLMUnavailable, err)
}
