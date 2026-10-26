package domain

import (
	"strings"

	"github.com/google/uuid"
)

const (
	slugMaxLength = 48
	slugSuffixLen = 4
)

var translitTable = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh",
	'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o",
	'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "h", 'ц': "ts",
	'ч': "ch", 'ш': "sh", 'щ': "sch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

func NewGameSlug(source string, id uuid.UUID) string {
	base := slugify(source)
	suffix := strings.ReplaceAll(id.String(), "-", "")[:slugSuffixLen]

	if base == "" {
		base = "game"
	}

	return base + "-" + suffix
}

func slugify(source string) string {
	var b strings.Builder

	for _, r := range strings.ToLower(source) {
		if mapped, ok := translitTable[r]; ok {
			b.WriteString(mapped)

			continue
		}

		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)

			continue
		}

		if r == ' ' || r == '-' || r == '_' {
			b.WriteByte('-')
		}
	}

	parts := make([]string, 0, 8)

	for _, part := range strings.Split(b.String(), "-") {
		if part == "" {
			continue
		}

		parts = append(parts, part)
	}

	slug := strings.Join(parts, "-")
	if len(slug) > slugMaxLength {
		slug = strings.Trim(slug[:slugMaxLength], "-")
	}

	return slug
}
