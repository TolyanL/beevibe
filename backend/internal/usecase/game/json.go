package game

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TolyanL/beevibe/internal/domain"
)

func extractJSONObject(text string) (string, error) {
	trimmed := strings.TrimSpace(text)

	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)

		if end := strings.Index(trimmed, "```"); end >= 0 {
			trimmed = trimmed[:end]
		}
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")

	if start < 0 || end < start {
		return "", fmt.Errorf("%w: json object not found in response", domain.ErrInvalidInput)
	}

	return trimmed[start : end+1], nil
}

func marshalSpec(spec domain.GameSpec) (string, error) {
	raw, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("json.Marshal: %w", err)
	}

	return string(raw), nil
}
