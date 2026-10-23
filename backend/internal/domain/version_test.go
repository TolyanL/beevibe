package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewVersionRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	gameID := uuid.New()

	tests := []struct {
		name   string
		number int
		html   string
	}{
		{name: "zero number", number: 0, html: "<html></html>"},
		{name: "negative number", number: -1, html: "<html></html>"},
		{name: "empty html", number: 1, html: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewVersion(gameID, tt.number, tt.html, VersionSourceGenerate)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("NewVersion() error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestNewVersionKeepsFields(t *testing.T) {
	t.Parallel()

	gameID := uuid.New()

	version, err := NewVersion(gameID, 3, "<html></html>", VersionSourcePatch)
	if err != nil {
		t.Fatalf("NewVersion() error = %v, want nil", err)
	}

	if version.GameID != gameID || version.Number != 3 || version.Source != VersionSourcePatch {
		t.Errorf("version = %+v, want game %s number 3 source patch", version, gameID)
	}

	if version.ID == uuid.Nil {
		t.Error("version id = nil, want uuid")
	}

	if version.CreatedAt.IsZero() {
		t.Error("created at = zero, want timestamp")
	}
}

func TestNewRatingAcceptsKnownValues(t *testing.T) {
	t.Parallel()

	gameID := uuid.New()
	userID := uuid.New()

	for _, value := range []RatingValue{RatingCool, RatingBad} {
		rating, err := NewRating(gameID, userID, value)
		if err != nil {
			t.Fatalf("NewRating() error = %v, want nil", err)
		}

		if rating.Value != value {
			t.Errorf("value = %s, want %s", rating.Value, value)
		}
	}
}

func TestNewRatingRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	_, err := NewRating(uuid.New(), uuid.New(), RatingValue("awesome"))
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewRating() error = %v, want ErrInvalidInput", err)
	}
}

func TestNewTokenOpRejectsZeroDelta(t *testing.T) {
	t.Parallel()

	_, err := NewTokenOp(uuid.New(), nil, 0, ReasonGenerate)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewTokenOp() error = %v, want ErrInvalidInput", err)
	}
}

func TestNewTokenOpKeepsGameLink(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	gameID := uuid.New()

	op, err := NewTokenOp(userID, &gameID, -320, ReasonGenerate)
	if err != nil {
		t.Fatalf("NewTokenOp() error = %v, want nil", err)
	}

	if op.GameID == nil || *op.GameID != gameID {
		t.Errorf("game id = %v, want %s", op.GameID, gameID)
	}

	if op.Delta != -320 || op.Reason != ReasonGenerate {
		t.Errorf("op = %+v, want delta -320 reason generate", op)
	}
}

func TestStatusPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    Status
		terminal  bool
		active    bool
		stringVal string
	}{
		{name: "ready is terminal", status: StatusReady, terminal: true, stringVal: "ready"},
		{name: "failed is terminal", status: StatusFailed, terminal: true, stringVal: "failed"},
		{name: "coding is active", status: StatusCoding, active: true, stringVal: "coding"},
		{name: "queued is active", status: StatusQueued, active: true, stringVal: "queued"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.terminal)
			}

			if got := tt.status.IsActive(); got != tt.active {
				t.Errorf("IsActive() = %v, want %v", got, tt.active)
			}

			if got := tt.status.String(); got != tt.stringVal {
				t.Errorf("String() = %q, want %q", got, tt.stringVal)
			}
		})
	}
}
