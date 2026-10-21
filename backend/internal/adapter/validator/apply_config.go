package validator

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

type configField struct {
	keys  []string
	value string
}

type edit struct {
	start int
	end   int
	text  string
}

// NOTE: правка выполняется без обращения к LLM: существующие ключи заменяются
// на месте, отсутствующие добавляются в конец блока.
func (v *Validator) ApplyConfig(ctx context.Context, html string, cfg domain.GameConfig) (string, error) {
	ctx, span := tracer.Start(ctx, "adapter validator ApplyConfig")
	defer span.End()

	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("validator.ApplyConfig: %w", err)
	}

	updated, err := applyConfig(html, cfg)
	if err != nil {
		return "", fmt.Errorf("validator.ApplyConfig: %w", err)
	}

	return updated, nil
}

func applyConfig(html string, cfg domain.GameConfig) (string, error) {
	start, end, ok := findConfigObject(html)
	if !ok {
		return "", fmt.Errorf("%w: config block not found", domain.ErrValidationFailed)
	}

	src := html[start : end+1]

	obj, err := parseJSObject(src)
	if err != nil {
		return "", fmt.Errorf("%w: %w", domain.ErrValidationFailed, err)
	}

	fields := []configField{
		{keys: []string{"tempo"}, value: strconv.Itoa(int(cfg.Tempo))},
		{keys: []string{"palette"}, value: renderPalette(cfg.Palette)},
		{keys: []string{"mapSize", "map_size"}, value: strconv.Quote(string(cfg.MapSize))},
		{keys: []string{"difficulty"}, value: strconv.Quote(cfg.Difficulty)},
		{keys: []string{"controls", "control"}, value: strconv.Quote(cfg.Controls)},
		{keys: []string{"style"}, value: strconv.Quote(string(cfg.Style))},
	}

	edits := make([]edit, 0, len(fields))
	missing := make([]configField, 0, len(fields))

	for _, field := range fields {
		member := obj.member(field.keys...)
		if member == nil {
			missing = append(missing, field)

			continue
		}

		edits = append(edits, edit{start: member.value.start, end: member.value.end, text: field.value})
	}

	if len(missing) > 0 {
		text, at := renderMissing(src, obj, missing)

		edits = append(edits, edit{start: at, end: at, text: text})
	}

	return html[:start] + applyEdits(src, edits) + html[end+1:], nil
}

func renderMissing(src string, obj *jsValue, missing []configField) (string, int) {
	multiline := strings.ContainsRune(src[obj.start:obj.end], '\n')

	parts := make([]string, 0, len(missing))
	for _, field := range missing {
		parts = append(parts, field.keys[0]+": "+field.value)
	}

	separator := ", "
	if multiline {
		separator = "\n" + memberIndent(src, obj)
	}

	body := strings.Join(parts, separator)

	if len(obj.members) == 0 {
		if multiline {
			return "\n" + memberIndent(src, obj) + body + "\n" + lineIndent(src, obj.start), obj.start + 1
		}

		return " " + body + " ", obj.start + 1
	}

	last := obj.members[len(obj.members)-1].value

	at := last.end
	for at < len(src) && isSpace(src[at]) {
		at++
	}

	if at < len(src) && src[at] == ',' {
		if multiline {
			return "\n" + memberIndent(src, obj) + body, at + 1
		}

		return " " + body, at + 1
	}

	if multiline {
		return ",\n" + memberIndent(src, obj) + body, last.end
	}

	return ", " + body, last.end
}

// NOTE: отступ членов — отступ последнего члена, иначе отступ открывающей
// скобки плюс два пробела.
func memberIndent(src string, obj *jsValue) string {
	if len(obj.members) > 0 {
		indent := lineIndent(src, obj.members[len(obj.members)-1].value.start)
		if indent != "" {
			return indent
		}
	}

	return lineIndent(src, obj.start) + "  "
}

// NOTE: правки применяются с конца к началу, чтобы смещения не «поехали».
func applyEdits(src string, edits []edit) string {
	slices.SortFunc(edits, func(a, b edit) int {
		return cmp.Compare(b.start, a.start)
	})

	out := src

	for _, e := range edits {
		out = out[:e.start] + e.text + out[e.end:]
	}

	return out
}

func renderPalette(palette []string) string {
	values := make([]string, 0, len(palette))

	for _, color := range palette {
		values = append(values, strconv.Quote(color))
	}

	return "[" + strings.Join(values, ", ") + "]"
}

func lineIndent(src string, pos int) string {
	if pos > len(src) {
		pos = len(src)
	}

	start := strings.LastIndexByte(src[:pos], '\n') + 1

	end := start
	for end < len(src) && (src[end] == ' ' || src[end] == '\t') {
		end++
	}

	return src[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
