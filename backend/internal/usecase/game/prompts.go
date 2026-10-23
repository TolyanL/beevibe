package game

import (
	"fmt"
	"strings"

	"github.com/TolyanL/beevibe/internal/domain"
)

func buildAnalyzePrompt(game domain.Game) string {
	var b strings.Builder

	b.WriteString("User request:\n")
	b.WriteString(game.Prompt)
	b.WriteString("\n\nParameters:\n")
	fmt.Fprintf(&b, "- genre: %s\n", genreLabel(game.Params.Genre))
	fmt.Fprintf(&b, "- tempo: %d (1 slow .. 10 fast)\n", game.Params.Tempo)
	fmt.Fprintf(&b, "- map size: %s\n", game.Params.MapSize)
	fmt.Fprintf(&b, "- style: %s\n", game.Params.Style)
	b.WriteString("\nThe free text of the request has priority over the parameters.")

	return b.String()
}

func buildCodePrompt(p *pipeline) string {
	spec, err := marshalSpec(p.spec)
	if err != nil {
		spec = ""
	}

	var b strings.Builder

	b.WriteString("Game spec:\n")
	b.WriteString(spec)
	b.WriteString("\n\nImplementation plan:\n")
	b.WriteString(p.plan)
	writeCanonicalConfig(&b, p)

	return b.String()
}

func buildReviewPrompt(p *pipeline) string {
	spec, err := marshalSpec(p.spec)
	if err != nil {
		spec = ""
	}

	var b strings.Builder

	b.WriteString("Game spec:\n")
	b.WriteString(spec)
	b.WriteString("\n\nGame code:\n")
	b.WriteString(p.code)

	return b.String()
}

func buildAssemblePrompt(p *pipeline) string {
	spec, err := marshalSpec(p.spec)
	if err != nil {
		spec = ""
	}

	var b strings.Builder

	b.WriteString("Game spec:\n")
	b.WriteString(spec)
	b.WriteString("\n\nGame code:\n")
	b.WriteString(p.code)
	b.WriteString("\n\nAssets:\n")
	b.WriteString(p.assets)
	writeCanonicalConfig(&b, p)

	return b.String()
}

// NOTE: спека несёт описательный стиль ("crisp pixel-art ..."), а CONFIG-блок
// требует enum-значение; без канонических значений из запроса модель копирует
// описание в поле style и финальная проверка конфига падает.
func writeCanonicalConfig(b *strings.Builder, p *pipeline) {
	block := canonicalConfigBlock(p.game.Params)
	if block == "" {
		return
	}

	b.WriteString("\n\nCanonical CONFIG values from the game request — copy them into the CONFIG block exactly as given:\n")
	b.WriteString(block)
	b.WriteString("The SPEC text describing the look is a human-readable description; never copy it into CONFIG fields.\n")
}

func canonicalConfigBlock(params domain.GameParams) string {
	var b strings.Builder

	if params.Tempo >= domain.TempoSlow && params.Tempo <= domain.TempoFast {
		fmt.Fprintf(&b, "- tempo: %d (integer %d..%d)\n", params.Tempo, domain.TempoSlow, domain.TempoFast)
	}

	if params.MapSize != "" {
		fmt.Fprintf(&b, "- mapSize: %q (allowed: %s)\n", params.MapSize, quotedValues(domain.MapSizeSmall, domain.MapSizeMedium, domain.MapSizeLarge))
	}

	if params.Style != "" {
		fmt.Fprintf(&b, "- style: %q (allowed: %s)\n", params.Style, quotedValues(domain.StylePixel, domain.StyleCartoon, domain.StyleNeon, domain.StyleMinimal))
	}

	return b.String()
}

func quotedValues[T ~string](values ...T) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}

	return strings.Join(quoted, ", ")
}

func buildRepairPrompt(html string, issues []domain.Issue) string {
	var b strings.Builder

	b.WriteString("Current game HTML:\n")
	b.WriteString(html)
	b.WriteString("\n\nIssues to fix:\n")
	b.WriteString(joinIssues(issues))

	return b.String()
}

func buildPatchPrompt(html, text, history string) string {
	var b strings.Builder

	b.WriteString("Current game HTML:\n")
	b.WriteString(html)

	if history != "" {
		b.WriteString("\n\nPrevious edits:\n")
		b.WriteString(history)
	}

	b.WriteString("\n\nRequested change:\n")
	b.WriteString(text)

	return b.String()
}

func genreLabel(genre domain.Genre) string {
	if genre == "" {
		return "not specified"
	}

	return string(genre)
}
