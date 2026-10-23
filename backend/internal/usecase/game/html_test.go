package game

import "testing"

func TestExtractHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "fenced html block",
			text: "```html\n<!DOCTYPE html>\n<html lang=\"ru\"></html>\n```",
			want: "<!DOCTYPE html>\n<html lang=\"ru\"></html>",
		},
		{
			name: "fenced block without language",
			text: "```\n<!DOCTYPE html>\n```",
			want: "<!DOCTYPE html>",
		},
		{
			name: "plain document",
			text: "<!DOCTYPE html>\n<html lang=\"ru\"></html>",
			want: "<!DOCTYPE html>\n<html lang=\"ru\"></html>",
		},
		{
			name: "prose around document",
			text: "Готово:\n```html\n<!DOCTYPE html>\n```\nСпасибо!",
			want: "<!DOCTYPE html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := extractHTML(tt.text); got != tt.want {
				t.Errorf("extractHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}
