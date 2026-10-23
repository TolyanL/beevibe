package validator

import (
	"context"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestValidatorCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		html        string
		wantCodes   []string
		requireNode bool
	}{
		{
			name: "valid game",
			html: validGameHTML,
		},
		{
			name:      "external url",
			html:      gameWithExternalURL(),
			wantCodes: []string{IssueExternalURL},
		},
		{
			name:      "no config block",
			html:      gameWithoutConfig(),
			wantCodes: []string{IssueNoConfig},
		},
		{
			name:      "forbidden api",
			html:      gameWithForbiddenAPI(),
			wantCodes: []string{IssueForbiddenAPI},
		},
		{
			name:        "broken javascript",
			html:        gameWithBrokenJS(),
			wantCodes:   []string{IssueJSSyntax},
			requireNode: true,
		},
		{
			name:      "empty document",
			html:      "   \n\t ",
			wantCodes: []string{IssueEmptyDocument},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.requireNode {
				if _, err := exec.LookPath("node"); err != nil {
					t.Skip("node is not available")
				}
			}

			v := New(Config{})

			issues, err := v.Check(context.Background(), tt.html)
			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}

			codes := issueCodes(issues)
			if !slices.Equal(codes, tt.wantCodes) {
				t.Errorf("Check() codes = %v, want %v", codes, tt.wantCodes)
			}
		})
	}
}

func TestValidatorCheckValidGameHasNoIssues(t *testing.T) {
	t.Parallel()

	v := New(Config{})

	issues, err := v.Check(context.Background(), validGameHTML)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if issues != nil {
		t.Errorf("Check() issues = %v, want nil", issues)
	}
}

func TestValidatorCheckReportsLine(t *testing.T) {
	t.Parallel()

	v := New(Config{})

	issues, err := v.Check(context.Background(), gameWithForbiddenAPI())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Check() issues = %v, want one issue", issues)
	}

	if issues[0].Line == 0 {
		t.Errorf("Check() line = %d, want the line of the forbidden API", issues[0].Line)
	}
}

func TestValidatorCheckSkipsSyntaxWithoutNode(t *testing.T) {
	t.Parallel()

	v := New(Config{NodePath: filepath.Join(t.TempDir(), "node")})

	issues, err := v.Check(context.Background(), gameWithBrokenJS())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if slices.Contains(issueCodes(issues), IssueJSSyntax) {
		t.Errorf("Check() codes = %v, want no js_syntax issue", issueCodes(issues))
	}
}

func TestValidatorCheckReportsSyntaxError(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not available")
	}

	v := New(Config{})

	issues, err := v.Check(context.Background(), gameWithBrokenJS())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	found := false

	for _, issue := range issues {
		if issue.Code != IssueJSSyntax {
			continue
		}

		found = true

		if !strings.Contains(issue.Message, "SyntaxError") {
			t.Errorf("Check() message = %q, want the node output", issue.Message)
		}

		if issue.Line == 0 {
			t.Error("Check() line = 0, want the line of the broken script")
		}
	}

	if !found {
		t.Fatalf("Check() codes = %v, want %s", issueCodes(issues), IssueJSSyntax)
	}
}

func issueCodes(issues []domain.Issue) []string {
	codes := make([]string, 0, len(issues))

	for _, issue := range issues {
		codes = append(codes, issue.Code)
	}

	return codes
}
