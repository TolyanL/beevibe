package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestGameStartStageTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		from    Status
		to      Status
		wantErr bool
	}{
		{name: "queued to analyzing", from: StatusQueued, to: StatusAnalyzing},
		{name: "analyzing to planning", from: StatusAnalyzing, to: StatusPlanning},
		{name: "planning to coding", from: StatusPlanning, to: StatusCoding},
		{name: "coding to assets", from: StatusCoding, to: StatusAssets},
		{name: "assets to reviewing", from: StatusAssets, to: StatusReviewing},
		{name: "reviewing to assembling", from: StatusReviewing, to: StatusAssembling},
		{name: "reviewing back to coding", from: StatusReviewing, to: StatusCoding},
		{name: "assembling to checking", from: StatusAssembling, to: StatusChecking},
		{name: "checking to ready", from: StatusChecking, to: StatusReady},
		{name: "ready to patching", from: StatusReady, to: StatusPatching},
		{name: "patching to checking", from: StatusPatching, to: StatusChecking},
		{name: "queued to coding is illegal", from: StatusQueued, to: StatusCoding, wantErr: true},
		{name: "ready to analyzing is illegal", from: StatusReady, to: StatusAnalyzing, wantErr: true},
		{name: "failed to coding is illegal", from: StatusFailed, to: StatusCoding, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			game := Game{Status: tt.from}

			err := game.StartStage(tt.to)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidTransition) {
					t.Fatalf("StartStage() error = %v, want ErrInvalidTransition", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("StartStage() error = %v, want nil", err)
			}

			if game.Status != tt.to {
				t.Fatalf("status = %s, want %s", game.Status, tt.to)
			}
		})
	}
}

func TestGameFail(t *testing.T) {
	t.Parallel()

	game := Game{Status: StatusCoding}

	if err := game.Fail("llm unavailable"); err != nil {
		t.Fatalf("Fail() error = %v, want nil", err)
	}

	if game.Status != StatusFailed {
		t.Fatalf("status = %s, want %s", game.Status, StatusFailed)
	}

	if game.FailReason != "llm unavailable" {
		t.Fatalf("fail reason = %q, want %q", game.FailReason, "llm unavailable")
	}

	if err := game.Fail("second time"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Fail() error = %v, want ErrInvalidTransition", err)
	}
}

func TestGameIsBusy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status Status
		want   bool
	}{
		{status: StatusQueued, want: true},
		{status: StatusCoding, want: true},
		{status: StatusReady, want: false},
		{status: StatusFailed, want: false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()

			game := Game{Status: tt.status}
			if game.IsBusy() != tt.want {
				t.Fatalf("IsBusy() = %v, want %v", game.IsBusy(), tt.want)
			}
		})
	}
}

func TestGameIsOwner(t *testing.T) {
	t.Parallel()

	owner := uuid.New()
	other := uuid.New()

	game := Game{OwnerID: owner}

	if !game.IsOwner(owner) {
		t.Fatal("IsOwner() = false, want true for owner")
	}

	if game.IsOwner(other) {
		t.Fatal("IsOwner() = true, want false for other user")
	}
}

func TestNewGame(t *testing.T) {
	t.Parallel()

	if _, err := NewGame(uuid.New(), "", GameParams{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewGame() error = %v, want ErrInvalidInput", err)
	}

	game, err := NewGame(uuid.New(), "змейка для бабушки", GameParams{Genre: GenreSnake})
	if err != nil {
		t.Fatalf("NewGame() error = %v, want nil", err)
	}

	if game.Status != StatusQueued {
		t.Fatalf("status = %s, want %s", game.Status, StatusQueued)
	}
}
