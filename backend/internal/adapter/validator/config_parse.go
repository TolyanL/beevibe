package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type jsKind int

const (
	jsObject jsKind = iota
	jsArray
	jsString
	jsNumber
	jsBool
	jsNull
	jsUnknown
)

// NOTE: start/end — границы значения в исходном тексте: они нужны для
// точечной правки блока CONFIG.
type jsValue struct {
	kind    jsKind
	str     string
	num     float64
	boolean bool
	members []jsMember
	elems   []*jsValue
	start   int
	end     int
}

type jsMember struct {
	key   string
	value *jsValue
}

type jsParseError struct {
	offset int
	reason string
}

func (e *jsParseError) Error() string {
	return fmt.Sprintf("js literal at offset %d: %s", e.offset, e.reason)
}

// NOTE: возвращаются индексы открывающей и закрывающей скобок:
// текст объекта — html[start:end+1].
func findConfigObject(html string) (int, int, bool) {
	decl := regexp.MustCompile(`\b(?:const|let|var)\s+CONFIG\s*=\s*\{`)

	loc := decl.FindStringIndex(html)
	if loc == nil {
		return 0, 0, false
	}

	open := loc[1] - 1

	end, ok := matchBrace(html, open)
	if !ok {
		return 0, 0, false
	}

	return open, end, true
}

func (v *jsValue) member(names ...string) *jsMember {
	for _, name := range names {
		for i := range v.members {
			if v.members[i].key == name {
				return &v.members[i]
			}
		}
	}

	for _, name := range names {
		for i := range v.members {
			if strings.EqualFold(v.members[i].key, name) {
				return &v.members[i]
			}
		}
	}

	return nil
}

type jsParser struct {
	src string
	pos int
}

// NOTE: src должен начинаться с '{'.
func parseJSObject(src string) (*jsValue, error) {
	p := &jsParser{src: src}
	p.skipSpace()

	if p.pos >= len(p.src) || p.src[p.pos] != '{' {
		return nil, p.errf("object literal expected")
	}

	return p.parseObject()
}

func (p *jsParser) errf(format string, args ...any) error {
	return &jsParseError{offset: p.pos, reason: fmt.Sprintf(format, args...)}
}

func (p *jsParser) parseObject() (*jsValue, error) {
	start := p.pos
	p.pos++

	value := &jsValue{kind: jsObject, start: start}

	for {
		p.skipSpace()

		if p.pos >= len(p.src) {
			return nil, p.errf("unexpected end of object literal")
		}

		if p.src[p.pos] == '}' {
			value.end = p.pos + 1
			p.pos++

			return value, nil
		}

		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}

		p.skipSpace()

		if p.pos >= len(p.src) || p.src[p.pos] != ':' {
			return nil, p.errf("expected ':' after key %q", key)
		}

		p.pos++

		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		value.members = append(value.members, jsMember{key: key, value: item})

		p.skipSpace()

		if p.pos < len(p.src) && p.src[p.pos] == ',' {
			p.pos++
		}
	}
}

func (p *jsParser) parseArray() (*jsValue, error) {
	start := p.pos
	p.pos++

	value := &jsValue{kind: jsArray, start: start}

	for {
		p.skipSpace()

		if p.pos >= len(p.src) {
			return nil, p.errf("unexpected end of array literal")
		}

		if p.src[p.pos] == ']' {
			value.end = p.pos + 1
			p.pos++

			return value, nil
		}

		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		value.elems = append(value.elems, item)

		p.skipSpace()

		if p.pos < len(p.src) && p.src[p.pos] == ',' {
			p.pos++
		}
	}
}

func (p *jsParser) parseValue() (*jsValue, error) {
	p.skipSpace()

	if p.pos >= len(p.src) {
		return nil, p.errf("value expected")
	}

	start := p.pos

	switch c := p.src[p.pos]; {
	case c == '{':
		return p.parseObject()
	case c == '[':
		return p.parseArray()
	case c == '"' || c == '\'':
		text, err := p.parseString()
		if err != nil {
			return nil, err
		}

		return &jsValue{kind: jsString, str: text, start: start, end: p.pos}, nil
	case c == '-' || c == '+' || isDigit(c):
		num, err := p.parseNumber()
		if err != nil {
			return nil, err
		}

		return &jsValue{kind: jsNumber, num: num, start: start, end: p.pos}, nil
	case strings.HasPrefix(p.src[p.pos:], "true"):
		p.pos += len("true")

		return &jsValue{kind: jsBool, boolean: true, start: start, end: p.pos}, nil
	case strings.HasPrefix(p.src[p.pos:], "false"):
		p.pos += len("false")

		return &jsValue{kind: jsBool, boolean: false, start: start, end: p.pos}, nil
	case strings.HasPrefix(p.src[p.pos:], "null"):
		p.pos += len("null")

		return &jsValue{kind: jsNull, start: start, end: p.pos}, nil
	}

	return p.parseUnknown()
}

// NOTE: значение-выражение берётся как есть, чтобы его можно было заменить целиком.
func (p *jsParser) parseUnknown() (*jsValue, error) {
	start := p.pos
	depth := 0

	for p.pos < len(p.src) {
		switch c := p.src[p.pos]; c {
		case '"', '\'', '`':
			next, ok := skipString(p.src, p.pos)
			if !ok {
				return nil, p.errf("unterminated string literal")
			}

			p.pos = next

			continue
		case '/':
			if next, ok := skipComment(p.src, p.pos); ok {
				p.pos = next

				continue
			}
		case '(', '[', '{':
			depth++
		case ')', ']':
			depth--
		case '}':
			if depth == 0 {
				return &jsValue{kind: jsUnknown, str: p.src[start:p.pos], start: start, end: p.pos}, nil
			}

			depth--
		case ',':
			if depth == 0 {
				return &jsValue{kind: jsUnknown, str: p.src[start:p.pos], start: start, end: p.pos}, nil
			}
		}

		p.pos++
	}

	return nil, p.errf("unexpected end of value")
}

func (p *jsParser) parseKey() (string, error) {
	if p.pos >= len(p.src) {
		return "", p.errf("unexpected end of object literal")
	}

	if c := p.src[p.pos]; c == '"' || c == '\'' {
		return p.parseString()
	}

	start := p.pos

	for p.pos < len(p.src) && isIdentChar(p.src[p.pos]) {
		p.pos++
	}

	if p.pos == start {
		return "", p.errf("object key expected")
	}

	return p.src[start:p.pos], nil
}

func (p *jsParser) parseString() (string, error) {
	quote := p.src[p.pos]
	p.pos++

	var b strings.Builder

	for p.pos < len(p.src) {
		c := p.src[p.pos]

		if c == '\\' {
			p.pos++

			if p.pos >= len(p.src) {
				return "", p.errf("unterminated escape sequence")
			}

			r, err := p.parseEscape()
			if err != nil {
				return "", err
			}

			b.WriteRune(r)

			continue
		}

		if c == quote {
			p.pos++

			return b.String(), nil
		}

		b.WriteByte(c)
		p.pos++
	}

	return "", p.errf("unterminated string literal")
}

func (p *jsParser) parseEscape() (rune, error) {
	c := p.src[p.pos]
	p.pos++

	switch c {
	case 'n':
		return '\n', nil
	case 't':
		return '\t', nil
	case 'r':
		return '\r', nil
	case 'b':
		return '\b', nil
	case 'f':
		return '\f', nil
	case 'v':
		return '\v', nil
	case '0':
		return 0, nil
	case 'x':
		return p.parseHex(2)
	case 'u':
		return p.parseHex(4)
	}

	return rune(c), nil
}

func (p *jsParser) parseHex(width int) (rune, error) {
	if p.pos+width > len(p.src) {
		return 0, p.errf("incomplete escape sequence")
	}

	digits := p.src[p.pos : p.pos+width]
	p.pos += width

	value, err := strconv.ParseUint(digits, 16, 32)
	if err != nil {
		return 0, p.errf("invalid escape sequence %q", digits)
	}

	return rune(value), nil
}

func (p *jsParser) parseNumber() (float64, error) {
	start := p.pos

	if c := p.src[p.pos]; c == '-' || c == '+' {
		p.pos++
	}

	for p.pos < len(p.src) && isDigit(p.src[p.pos]) {
		p.pos++
	}

	if p.pos < len(p.src) && p.src[p.pos] == '.' {
		p.pos++

		for p.pos < len(p.src) && isDigit(p.src[p.pos]) {
			p.pos++
		}
	}

	if p.pos < len(p.src) && (p.src[p.pos] == 'e' || p.src[p.pos] == 'E') {
		p.pos++

		if p.pos < len(p.src) && (p.src[p.pos] == '-' || p.src[p.pos] == '+') {
			p.pos++
		}

		for p.pos < len(p.src) && isDigit(p.src[p.pos]) {
			p.pos++
		}
	}

	text := p.src[start:p.pos]
	if text == "" {
		return 0, p.errf("number expected")
	}

	num, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, p.errf("invalid number %q", text)
	}

	return num, nil
}

func (p *jsParser) skipSpace() {
	for p.pos < len(p.src) {
		c := p.src[p.pos]

		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p.pos++

			continue
		}

		if c == '/' {
			if next, ok := skipComment(p.src, p.pos); ok {
				p.pos = next

				continue
			}
		}

		return
	}
}

func matchBrace(src string, open int) (int, bool) {
	if open < 0 || open >= len(src) || src[open] != '{' {
		return 0, false
	}

	depth := 0

	for i := open; i < len(src); {
		switch src[i] {
		case '"', '\'', '`':
			next, ok := skipString(src, i)
			if !ok {
				return 0, false
			}

			i = next

			continue
		case '/':
			if next, ok := skipComment(src, i); ok {
				i = next

				continue
			}
		case '{':
			depth++
		case '}':
			depth--

			if depth == 0 {
				return i, true
			}
		}

		i++
	}

	return 0, false
}

func skipString(src string, i int) (int, bool) {
	quote := src[i]

	for j := i + 1; j < len(src); {
		switch src[j] {
		case '\\':
			j += 2

			continue
		case quote:
			return j + 1, true
		}

		j++
	}

	return 0, false
}

func skipComment(src string, i int) (int, bool) {
	if i+1 >= len(src) {
		return 0, false
	}

	switch src[i+1] {
	case '/':
		j := i + 2

		for j < len(src) && src[j] != '\n' {
			j++
		}

		return j, true
	case '*':
		j := i + 2

		for j+1 < len(src) {
			if src[j] == '*' && src[j+1] == '/' {
				return j + 2, true
			}

			j++
		}

		return 0, false
	}

	return 0, false
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentChar(c byte) bool {
	return isLetter(c) || isDigit(c) || c == '_' || c == '$'
}
