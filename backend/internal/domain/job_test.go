package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewGameJobAcceptsPatchWithInstruction(t *testing.T) {
	t.Parallel()

	gameID := uuid.New()
	ownerID := uuid.New()

	job, err := NewGameJob(gameID, ownerID, JobKindPatch, "  сделай фон темнее  ")
	if err != nil {
		t.Fatalf("NewGameJob() error = %v, want nil", err)
	}

	if job.GameID != gameID || job.OwnerID != ownerID || job.Kind != JobKindPatch {
		t.Errorf("job = %+v, want patch job of game %s", job, gameID)
	}

	if job.Instruction != "сделай фон темнее" {
		t.Errorf("instruction = %q, want trimmed text", job.Instruction)
	}

	if job.ID == uuid.Nil || job.CreatedAt.IsZero() {
		t.Errorf("job = %+v, want generated id and timestamp", job)
	}
}

func TestNewGameJobRequiresPatchInstruction(t *testing.T) {
	t.Parallel()

	_, err := NewGameJob(uuid.New(), uuid.New(), JobKindPatch, "   ")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewGameJob() error = %v, want ErrInvalidInput", err)
	}
}

func TestNewGameJobAllowsFixWithoutInstruction(t *testing.T) {
	t.Parallel()

	job, err := NewGameJob(uuid.New(), uuid.New(), JobKindFix, "")
	if err != nil {
		t.Fatalf("NewGameJob() error = %v, want nil", err)
	}

	if job.Kind != JobKindFix || job.Instruction != "" {
		t.Errorf("job = %+v, want fix job without instruction", job)
	}
}

func TestNewGameJobRejectsUnknownKind(t *testing.T) {
	t.Parallel()

	_, err := NewGameJob(uuid.New(), uuid.New(), JobKind("rollback"), "текст")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewGameJob() error = %v, want ErrInvalidInput", err)
	}
}

func TestNewGameJobRejectsTooLongInstruction(t *testing.T) {
	t.Parallel()

	_, err := NewGameJob(uuid.New(), uuid.New(), JobKindPatch, strings.Repeat("а", jobInstructionMaxLength+1))
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewGameJob() error = %v, want ErrInvalidInput", err)
	}
}

func TestJobKindIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind JobKind
		want bool
	}{
		{kind: JobKindPatch, want: true},
		{kind: JobKindFix, want: true},
		{kind: "", want: false},
		{kind: "generate", want: false},
	}

	for _, tt := range tests {
		if got := tt.kind.IsValid(); got != tt.want {
			t.Errorf("JobKind(%q).IsValid() = %t, want %t", tt.kind, got, tt.want)
		}
	}
}
