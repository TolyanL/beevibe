package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestEffortValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		effort Effort
		want   bool
	}{
		{EffortDefault, true},
		{effortDefaultLiteral, true},
		{EffortNone, true},
		{EffortMinimal, true},
		{EffortLow, true},
		{EffortMedium, true},
		{EffortHigh, true},
		{EffortXHigh, true},
		{EffortMax, true},
		{"turbo", false},
		{"LOW", false},
	}

	for _, tt := range tests {
		if got := tt.effort.Valid(); got != tt.want {
			t.Errorf("Effort(%q).Valid() = %v, want %v", tt.effort, got, tt.want)
		}
	}
}

func TestChatBodyReasoning(t *testing.T) {
	t.Parallel()

	const legacyJSON = `{"model":"test-model",` +
		`"messages":[{"role":"system","content":"system prompt"},` +
		`{"role":"user","content":"user prompt"}],` +
		`"temperature":0.4}`

	tests := []struct {
		name       string
		dialect    string
		effort     Effort
		wantErr    bool
		wantNested bool
		wantTop    bool
		wantValue  string
		wantLegacy bool
	}{
		{
			name:       "openrouter default sends nothing",
			dialect:    ReasoningDialectOpenRouter,
			effort:     EffortDefault,
			wantLegacy: true,
		},
		{
			name:       "openrouter default literal sends nothing",
			dialect:    ReasoningDialectOpenRouter,
			effort:     effortDefaultLiteral,
			wantLegacy: true,
		},
		{
			name:       "openai default literal sends nothing",
			dialect:    ReasoningDialectOpenAI,
			effort:     effortDefaultLiteral,
			wantLegacy: true,
		},
		{
			name:       "empty dialect means openrouter",
			dialect:    "",
			effort:     EffortLow,
			wantNested: true,
			wantValue:  "low",
		},
		{
			name:       "openrouter nests effort",
			dialect:    ReasoningDialectOpenRouter,
			effort:     EffortHigh,
			wantNested: true,
			wantValue:  "high",
		},
		{
			name:      "openai sends top-level effort",
			dialect:   ReasoningDialectOpenAI,
			effort:    EffortNone,
			wantTop:   true,
			wantValue: "none",
		},
		{
			name:    "unknown effort is rejected",
			dialect: ReasoningDialectOpenRouter,
			effort:  "turbo",
			wantErr: true,
		},
		{
			name:    "unknown dialect with effort is rejected",
			dialect: "bogus",
			effort:  EffortLow,
			wantErr: true,
		},
		{
			name:       "unknown dialect without effort keeps legacy body",
			dialect:    "bogus",
			effort:     EffortDefault,
			wantLegacy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := New(Config{Model: "test-model", ReasoningDialect: tt.dialect})

			body, err := client.chatBody(Request{
				System:      "system prompt",
				Prompt:      "user prompt",
				Temperature: 0.4,
				Reasoning:   tt.effort,
			})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantLegacy {
				if string(body) != legacyJSON {
					t.Fatalf("body changed:\nwant %s\n got %s", legacyJSON, body)
				}

				return
			}

			var decoded map[string]json.RawMessage

			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			_, hasNested := decoded["reasoning"]
			_, hasTop := decoded["reasoning_effort"]

			if hasNested != tt.wantNested {
				t.Errorf("reasoning field presence = %v, want %v", hasNested, tt.wantNested)
			}

			if hasTop != tt.wantTop {
				t.Errorf("reasoning_effort field presence = %v, want %v", hasTop, tt.wantTop)
			}

			if tt.wantNested {
				var payload reasoningPayload
				if err := json.Unmarshal(decoded["reasoning"], &payload); err != nil {
					t.Fatalf("json.Unmarshal reasoning: %v", err)
				}

				if payload.Effort != tt.wantValue {
					t.Errorf("nested effort = %q, want %q", payload.Effort, tt.wantValue)
				}
			}

			if tt.wantTop {
				var effort string
				if err := json.Unmarshal(decoded["reasoning_effort"], &effort); err != nil {
					t.Fatalf("json.Unmarshal reasoning_effort: %v", err)
				}

				if effort != tt.wantValue {
					t.Errorf("top-level effort = %q, want %q", effort, tt.wantValue)
				}
			}
		})
	}
}

func TestCompleteDebugDump(t *testing.T) {
	t.Parallel()

	const (
		errorBody = `{"error":"upstream boom"}`
		okRaw     = `{"choices":[{"message":{"role":"assistant",` +
			`"content":"готово","reasoning_content":"думал"},` +
			`"finish_reason":"stop"}],"usage":{"prompt_tokens":11,` +
			`"completion_tokens":22,"completion_tokens_details":{"reasoning_tokens":7}}}`
	)

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(errorBody))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(okRaw))
	}))
	defer server.Close()

	debugDir := t.TempDir()

	client := New(Config{
		APIKey:           "test-key",
		Model:            "test-model",
		BaseURL:          server.URL,
		Timeout:          5 * time.Second,
		MaxRetries:       1,
		RetryBackoff:     time.Millisecond,
		ReasoningDialect: ReasoningDialectOpenAI,
		DebugDir:         debugDir,
	})

	response, err := client.Complete(t.Context(), Request{
		System:    "system prompt",
		Prompt:    "user prompt",
		Reasoning: EffortHigh,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.Text != "готово" {
		t.Fatalf("text = %q, want %q", response.Text, "готово")
	}

	entries, err := os.ReadDir(debugDir)
	if err != nil {
		t.Fatalf("os.ReadDir: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("dump dirs = %d, want 1", len(entries))
	}

	dumpDir := filepath.Join(debugDir, entries[0].Name())

	requestBody := readFile(t, filepath.Join(dumpDir, "attempt-1.req.json"))
	if !strings.Contains(string(requestBody), `"reasoning_effort":"high"`) {
		t.Errorf("request dump misses reasoning_effort: %s", requestBody)
	}

	firstResponse := readFile(t, filepath.Join(dumpDir, "attempt-1.resp.json"))
	if string(firstResponse) != errorBody {
		t.Errorf("first response dump = %s, want %s", firstResponse, errorBody)
	}

	var firstMeta debugMeta
	if err := json.Unmarshal(readFile(t, filepath.Join(dumpDir, "attempt-1.meta.json")), &firstMeta); err != nil {
		t.Fatalf("json.Unmarshal meta: %v", err)
	}

	if firstMeta.Attempt != 1 || firstMeta.Error == "" {
		t.Errorf("first meta = %+v, want attempt 1 with error", firstMeta)
	}

	if firstMeta.Effort != "high" || firstMeta.Dialect != ReasoningDialectOpenAI {
		t.Errorf("first meta effort/dialect = %q/%q", firstMeta.Effort, firstMeta.Dialect)
	}

	secondResponse := readFile(t, filepath.Join(dumpDir, "attempt-2.resp.json"))
	if string(secondResponse) != okRaw {
		t.Errorf("second response dump = %s, want %s", secondResponse, okRaw)
	}

	var secondMeta debugMeta
	if err := json.Unmarshal(readFile(t, filepath.Join(dumpDir, "attempt-2.meta.json")), &secondMeta); err != nil {
		t.Fatalf("json.Unmarshal meta: %v", err)
	}

	if secondMeta.Attempt != 2 || secondMeta.Error != "" {
		t.Errorf("second meta = %+v, want attempt 2 without error", secondMeta)
	}

	if secondMeta.ReasoningTokens != 7 {
		t.Errorf("reasoning tokens = %d, want 7", secondMeta.ReasoningTokens)
	}
}

func TestCompleteDebugDumpWriteFailureDoesNotFailCall(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(okBody))
	}))
	defer server.Close()

	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	client := New(Config{
		APIKey:  "test-key",
		Model:   "test-model",
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
		// Родитель пути — файл: MkdirAll гарантированно падает.
		DebugDir: filepath.Join(blocker, "debug"),
	})

	response, err := client.Complete(t.Context(), Request{System: "system", Prompt: "prompt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.Text == "" {
		t.Fatal("empty response text")
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%s): %v", path, err)
	}

	return data
}
