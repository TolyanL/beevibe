package validator

import (
	"reflect"
	"testing"
)

func TestParseJSObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		src     string
		want    map[string]any
		wantErr bool
	}{
		{
			name: "numbers, strings and booleans",
			src:  `{ "tempo": 5, "style": 'pixel', "on": true, "off": false }`,
			want: map[string]any{"tempo": 5.0, "style": "pixel", "on": true, "off": false},
		},
		{
			name: "nested object and array",
			src:  `{ "palette": ["#fff", "#000"], "levels": { "one": 1 } }`,
			want: map[string]any{"palette": []any{"#fff", "#000"}, "levels": map[string]any{"one": 1.0}},
		},
		{
			name: "escapes and trailing comma",
			src:  `{ "title": "a\"b\n", "tempo": 3, }`,
			want: map[string]any{"title": "a\"b\n", "tempo": 3.0},
		},
		{
			name: "comments are skipped",
			src:  "{ // comment\n \"tempo\": 5 /* block */ }",
			want: map[string]any{"tempo": 5.0},
		},
		{
			name:    "unterminated object",
			src:     `{ "tempo": 5`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			obj, err := parseJSObject(tt.src)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseJSObject() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("parseJSObject() error = %v", err)
			}

			got := map[string]any{}
			for _, member := range obj.members {
				got[member.key] = jsValueToAny(member.value)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseJSObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindConfigObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		html    string
		want    string
		wantErr bool
	}{
		{
			name: "const declaration",
			html: `const CONFIG = { "tempo": 5 };`,
			want: `{ "tempo": 5 }`,
		},
		{
			name: "braces inside strings",
			html: `const CONFIG = { "title": "a{b}c" };`,
			want: `{ "title": "a{b}c" }`,
		},
		{
			name: "nested object",
			html: `const CONFIG = { "levels": { "one": { "x": 1 } } };`,
			want: `{ "levels": { "one": { "x": 1 } } }`,
		},
		{
			name:    "no config block",
			html:    `const OTHER = { "tempo": 5 };`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			start, end, ok := findConfigObject(tt.html)
			if tt.wantErr {
				if ok {
					t.Fatalf("findConfigObject() found a block in %q", tt.html)
				}

				return
			}

			if !ok {
				t.Fatalf("findConfigObject() found no block in %q", tt.html)
			}

			if got := tt.html[start : end+1]; got != tt.want {
				t.Errorf("findConfigObject() = %q, want %q", got, tt.want)
			}
		})
	}
}

func jsValueToAny(value *jsValue) any {
	switch value.kind {
	case jsObject:
		out := map[string]any{}

		for _, member := range value.members {
			out[member.key] = jsValueToAny(member.value)
		}

		return out
	case jsArray:
		out := make([]any, 0, len(value.elems))

		for _, elem := range value.elems {
			out = append(out, jsValueToAny(elem))
		}

		return out
	case jsString:
		return value.str
	case jsNumber:
		return value.num
	case jsBool:
		return value.boolean
	default:
		return nil
	}
}
