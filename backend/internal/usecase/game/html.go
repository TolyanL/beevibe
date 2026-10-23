package game

import "strings"

func extractHTML(text string) string {
	trimmed := strings.TrimSpace(text)

	start := strings.Index(trimmed, "```")
	if start < 0 {
		return trimmed
	}

	trimmed = trimmed[start+len("```"):]
	trimmed = trimFenceLang(trimmed)

	if end := strings.Index(trimmed, "```"); end >= 0 {
		trimmed = trimmed[:end]
	}

	return strings.TrimSpace(trimmed)
}

func trimFenceLang(text string) string {
	line, rest, found := strings.Cut(text, "\n")
	if !found {
		return text
	}

	if !isLangTag(line) {
		return text
	}

	return rest
}

func isLangTag(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}

	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '+', r == '_':
		default:
			return false
		}
	}

	return true
}
