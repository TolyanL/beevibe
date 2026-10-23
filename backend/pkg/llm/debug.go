package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

type debugMeta struct {
	Attempt         int    `json:"attempt"`
	DurationMS      int64  `json:"duration_ms"`
	Error           string `json:"error,omitempty"`
	Effort          string `json:"effort"`
	Dialect         string `json:"dialect"`
	ReasoningTokens int    `json:"reasoning_tokens,omitempty"`
}

type debugDump struct {
	dir string
}

func (c *Client) newDebugDump() *debugDump {
	if c.cfg.DebugDir == "" {
		return nil
	}

	return &debugDump{
		dir: filepath.Join(c.cfg.DebugDir, fmt.Sprintf("%d", time.Now().UnixNano())),
	}
}

func (d *debugDump) attempt(
	number int,
	request, response []byte,
	duration time.Duration,
	attemptErr error,
	effort Effort,
	dialect string,
) {
	if d == nil {
		return
	}

	if err := os.MkdirAll(d.dir, 0o755); err != nil {
		log.Debug().Err(err).Str("dir", d.dir).Msg("llm debug dump: mkdir")

		return
	}

	errText := ""
	if attemptErr != nil {
		errText = attemptErr.Error()
	}

	meta := debugMeta{
		Attempt:         number,
		DurationMS:      duration.Milliseconds(),
		Error:           errText,
		Effort:          effort.value(),
		Dialect:         dialect,
		ReasoningTokens: reasoningTokens(response),
	}

	d.write(fmt.Sprintf("attempt-%d.req.json", number), request)
	d.write(fmt.Sprintf("attempt-%d.resp.json", number), response)
	d.writeJSON(fmt.Sprintf("attempt-%d.meta.json", number), meta)
}

func (d *debugDump) write(name string, data []byte) {
	if err := os.WriteFile(filepath.Join(d.dir, name), data, 0o644); err != nil {
		log.Debug().Err(err).Str("dir", d.dir).Str("file", name).Msg("llm debug dump: write")
	}
}

func (d *debugDump) writeJSON(name string, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Debug().Err(err).Str("dir", d.dir).Str("file", name).Msg("llm debug dump: marshal")

		return
	}

	d.write(name, data)
}

func reasoningTokens(raw []byte) int {
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return 0
	}

	return parsed.Usage.CompletionTokensDetails.ReasoningTokens
}
