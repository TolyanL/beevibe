package validator

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

const (
	IssueEmptyDocument = "empty_document"
	IssueExternalURL   = "external_url"
	IssueForbiddenAPI  = "forbidden_api"
	IssueJSSyntax      = "js_syntax"
	IssueNoConfig      = "no_config"
	IssueNoTouch       = "no_touch"
	IssueNoKeyboard    = "no_keyboard"
	IssueNoCanvas      = "no_canvas"
)

const jsSyntaxMessageLimit = 500

func (v *Validator) Check(ctx context.Context, html string) ([]domain.Issue, error) {
	ctx, span := tracer.Start(ctx, "adapter validator Check")
	defer span.End()

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("validator.Check: %w", err)
	}

	if !hasMarkup(html) {
		return []domain.Issue{{
			Code:    IssueEmptyDocument,
			Message: "document has no HTML markup",
		}}, nil
	}

	masked := maskNamespaceURLs(html)

	issues := externalResourceIssues(masked)
	issues = append(issues, forbiddenAPIIssues(masked)...)

	issue, ok, err := v.checkJSSyntax(ctx, html)
	if err != nil {
		return nil, fmt.Errorf("validator.Check: %w", err)
	}

	if ok {
		issues = append(issues, issue)
	}

	issues = append(issues, contractIssues(html)...)

	if len(issues) == 0 {
		return nil, nil
	}

	return issues, nil
}

type resourceCandidate struct {
	note string
	url  string
	at   int
}

// NOTE: игра должна быть самодостаточной: обращения к внешним ресурсам запрещены.
func externalResourceIssues(html string) []domain.Issue {
	candidates := resourceCandidates(html)

	seen := make(map[string]struct{}, len(candidates))
	issues := make([]domain.Issue, 0, len(candidates))

	for _, candidate := range candidates {
		key := candidate.note
		if candidate.url != "" {
			key = candidate.url
		}

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		message := candidate.note + " is not allowed"
		if candidate.url != "" {
			message = fmt.Sprintf("%s is not allowed: %s", candidate.note, snippet(candidate.url))
		}

		issues = append(issues, domain.Issue{
			Code:    IssueExternalURL,
			Message: message,
			Line:    lineAt(html, candidate.at),
		})
	}

	return issues
}

func resourceCandidates(html string) []resourceCandidate {
	var candidates []resourceCandidate

	candidates = append(candidates, tagURLCandidates(html, `(?i)<script\b[^>]*\ssrc\s*=\s*["']?([^"'\s>]*)`, "external script", false)...)
	candidates = append(candidates, tagURLCandidates(html, `(?i)<link\b[^>]*\shref\s*=\s*["']?([^"'\s>]*)`, "external stylesheet", true)...)

	absolute := regexp.MustCompile(`https?://[^\s"'<>()\[\]]+`)
	for _, loc := range absolute.FindAllStringIndex(html, -1) {
		candidates = append(candidates, resourceCandidate{
			note: "external URL",
			url:  html[loc[0]:loc[1]],
			at:   loc[0],
		})
	}

	relative := regexp.MustCompile(`(?:^|[\s"'(=,\[])//[a-zA-Z0-9][a-zA-Z0-9.-]*\.[a-zA-Z]{2,}`)
	for _, loc := range relative.FindAllStringIndex(html, -1) {
		start := loc[0]
		for start < loc[1] && html[start] != '/' {
			start++
		}

		candidates = append(candidates, resourceCandidate{
			note: "protocol-relative URL",
			url:  html[start:loc[1]],
			at:   start,
		})
	}

	candidates = append(candidates, importedURLCandidates(html)...)

	patterns := []struct {
		note string
		re   *regexp.Regexp
	}{
		{note: "fetch request", re: regexp.MustCompile(`\bfetch\s*\(`)},
		{note: "XMLHttpRequest", re: regexp.MustCompile(`\bXMLHttpRequest\b`)},
		{note: "web worker", re: regexp.MustCompile(`\bnew\s+Worker\s*\(`)},
		{note: "importScripts", re: regexp.MustCompile(`\bimportScripts\s*\(`)},
	}

	for _, pattern := range patterns {
		if loc := pattern.re.FindStringIndex(html); loc != nil {
			candidates = append(candidates, resourceCandidate{note: pattern.note, at: loc[0]})
		}
	}

	return candidates
}

// NOTE: externalOnly ограничивает поиск абсолютными адресами.
func tagURLCandidates(html, pattern, note string, externalOnly bool) []resourceCandidate {
	re := regexp.MustCompile(pattern)

	var candidates []resourceCandidate

	for _, loc := range re.FindAllStringSubmatchIndex(html, -1) {
		if len(loc) < 4 || loc[2] < 0 {
			continue
		}

		url := html[loc[2]:loc[3]]
		if url == "" {
			continue
		}

		if externalOnly && !isExternalURL(url) {
			continue
		}

		candidates = append(candidates, resourceCandidate{note: note, url: url, at: loc[0]})
	}

	return candidates
}

func importedURLCandidates(html string) []resourceCandidate {
	re := regexp.MustCompile(`\b(?:import\s*\(\s*|import\s+|from\s+)["']\s*((?:https?:)?//[^"'\s)]*)`)

	var candidates []resourceCandidate

	for _, loc := range re.FindAllStringSubmatchIndex(html, -1) {
		if len(loc) < 4 || loc[2] < 0 {
			continue
		}

		url := html[loc[2]:loc[3]]
		if url == "" {
			continue
		}

		candidates = append(candidates, resourceCandidate{note: "import from URL", url: url, at: loc[0]})
	}

	return candidates
}

func isExternalURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "//")
}

func forbiddenAPIIssues(html string) []domain.Issue {
	patterns := []struct {
		note string
		re   *regexp.Regexp
	}{
		{note: "localStorage", re: regexp.MustCompile(`\blocalStorage\b`)},
		{note: "sessionStorage", re: regexp.MustCompile(`\bsessionStorage\b`)},
		{note: "eval", re: regexp.MustCompile(`\beval\s*\(`)},
		{note: "new Function", re: regexp.MustCompile(`\bnew\s+Function\s*\(`)},
	}

	var issues []domain.Issue

	for _, pattern := range patterns {
		loc := pattern.re.FindStringIndex(html)
		if loc == nil {
			continue
		}

		issues = append(issues, domain.Issue{
			Code:    IssueForbiddenAPI,
			Message: fmt.Sprintf("%s is not allowed", pattern.note),
			Line:    lineAt(html, loc[0]),
		})
	}

	return issues
}

func contractIssues(html string) []domain.Issue {
	var issues []domain.Issue

	if _, _, ok := findConfigObject(html); !ok {
		issues = append(issues, domain.Issue{
			Code:    IssueNoConfig,
			Message: "game has no CONFIG block",
		})
	}

	lower := strings.ToLower(html)

	if !containsAny(lower, "touchstart", "touchend", "pointerdown") {
		issues = append(issues, domain.Issue{
			Code:    IssueNoTouch,
			Message: "game handles neither touch nor pointer events",
		})
	}

	if !containsAny(lower, "keydown", "keyup") {
		issues = append(issues, domain.Issue{
			Code:    IssueNoKeyboard,
			Message: "game handles no keyboard events",
		})
	}

	if !hasCanvas(lower) {
		issues = append(issues, domain.Issue{
			Code:    IssueNoCanvas,
			Message: "game has neither canvas nor game element",
		})
	}

	return issues
}

// NOTE: отсутствие node — не ошибка: проверка синтаксиса в этом случае пропускается.
func (v *Validator) checkJSSyntax(ctx context.Context, html string) (domain.Issue, bool, error) {
	if v.nodePath == "" {
		return domain.Issue{}, false, nil
	}

	for _, block := range extractScripts(html) {
		output, err := v.runNodeCheck(ctx, block.content)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return domain.Issue{}, false, nil
			}

			return domain.Issue{}, false, err
		}

		if output == "" {
			continue
		}

		line := block.line

		if nodeLine, ok := nodeErrorLine(output); ok {
			line += nodeLine - 1
		}

		return domain.Issue{
			Code:    IssueJSSyntax,
			Message: truncate(output, jsSyntaxMessageLimit),
			Line:    line,
		}, true, nil
	}

	return domain.Issue{}, false, nil
}

// NOTE: пустой вывод означает корректный синтаксис, непустой — текст ошибки node.
func (v *Validator) runNodeCheck(ctx context.Context, code string) (string, error) {
	dir, err := os.MkdirTemp("", "validator-js")
	if err != nil {
		return "", fmt.Errorf("os.MkdirTemp: %w", err)
	}

	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "script.mjs")

	if err := os.WriteFile(path, []byte(code), 0o600); err != nil {
		return "", fmt.Errorf("os.WriteFile: %w", err)
	}

	output, err := exec.CommandContext(ctx, v.nodePath, "--check", path).CombinedOutput()
	if err == nil {
		return "", nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(output), nil
	}

	return "", fmt.Errorf("exec.CommandContext: %w", err)
}

type scriptBlock struct {
	content string
	line    int
}

func extractScripts(html string) []scriptBlock {
	openTag := regexp.MustCompile(`(?i)<script\b[^>]*>`)
	closeTag := regexp.MustCompile(`(?i)</script\s*>`)
	srcAttr := regexp.MustCompile(`(?i)\ssrc\s*=`)

	var blocks []scriptBlock

	for _, loc := range openTag.FindAllStringIndex(html, -1) {
		close := closeTag.FindStringIndex(html[loc[1]:])
		if close == nil {
			break
		}

		contentStart := loc[1]
		contentEnd := loc[1] + close[0]

		if srcAttr.MatchString(html[loc[0]:loc[1]]) {
			continue
		}

		blocks = append(blocks, scriptBlock{
			content: html[contentStart:contentEnd],
			line:    lineAt(html, contentStart),
		})
	}

	return blocks
}

func nodeErrorLine(output string) (int, bool) {
	first := output

	if idx := strings.IndexByte(first, '\n'); idx >= 0 {
		first = first[:idx]
	}

	idx := strings.LastIndexByte(first, ':')
	if idx < 0 {
		return 0, false
	}

	line, err := strconv.Atoi(first[idx+1:])
	if err != nil {
		return 0, false
	}

	return line, true
}

// NOTE: xmlns не является внешним ресурсом: URL пространства имён маскируется,
// чтобы не считать его обращением наружу.
func maskNamespaceURLs(html string) string {
	namespace := regexp.MustCompile(`xmlns(:[a-zA-Z0-9_-]+)?\s*=\s*("|')https?://[^"']*("|')`)

	return namespace.ReplaceAllStringFunc(html, func(match string) string {
		return strings.Repeat(" ", len(match))
	})
}

func hasMarkup(html string) bool {
	for i := 0; i < len(html); i++ {
		if html[i] != '<' {
			continue
		}

		if i+1 >= len(html) {
			return false
		}

		switch next := html[i+1]; {
		case next == '!', next == '/', next == '?':
			return true
		case isLetter(next):
			return true
		}
	}

	return false
}

func hasCanvas(lower string) bool {
	return containsAny(lower,
		"<canvas",
		"getcontext(",
		"createelement(\"canvas\")",
		"createelement('canvas')",
		"id=\"game\"",
		"id='game'",
		"id=\"board\"",
		"id=\"app\"",
		"id=\"grid\"",
		"id=\"arena\"",
		"id=\"stage\"",
		"id=\"root\"",
	)
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}

	return false
}

func lineAt(html string, index int) int {
	if index > len(html) {
		index = len(html)
	}

	return 1 + strings.Count(html[:index], "\n")
}

func snippet(text string) string {
	return truncate(strings.TrimSpace(text), 80)
}

func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit-3]) + "..."
}
