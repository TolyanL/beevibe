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

// NOTE: the spec and the plan come from one answer, so the JSON object is cut out by brace
// matching instead of extractJSONObject, which would swallow braces that belong to the plan.
func parseSpecPlan(text string) (domain.GameSpec, string, error) {
	var spec domain.GameSpec

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return spec, "", fmt.Errorf("%w: empty response", domain.ErrInvalidInput)
	}

	raw, plan, ok := splitSpecBlock(trimmed)
	if !ok {
		return spec, "", fmt.Errorf("%w: spec object not found in response", domain.ErrInvalidInput)
	}

	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		return spec, "", fmt.Errorf("json.Unmarshal: %w", err)
	}

	return spec, strings.TrimSpace(plan), nil
}

func splitSpecBlock(text string) (string, string, bool) {
	if strings.HasPrefix(text, "```") {
		body := text
		body = strings.TrimPrefix(body, "```json")
		body = strings.TrimPrefix(body, "```JSON")
		body = strings.TrimPrefix(body, "```")
		body = strings.TrimSpace(body)

		if end := strings.Index(body, "```"); end >= 0 {
			return strings.TrimSpace(body[:end]), body[end+len("```"):], true
		}
	}

	start := strings.Index(text, "{")
	if start < 0 {
		return "", "", false
	}

	end, ok := matchBrace(text, start)
	if !ok {
		return "", "", false
	}

	return text[start : end+1], text[end+1:], true
}

func matchBrace(text string, start int) (int, bool) {
	depth := 0
	inString := false
	escaped := false

	for index := start; index < len(text); index++ {
		char := text[index]

		if inString {
			switch {
			case escaped:
				escaped = false
			case char == '\\':
				escaped = true
			case char == '"':
				inString = false
			}

			continue
		}

		switch char {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--

			if depth == 0 {
				return index, true
			}
		}
	}

	return 0, false
}
