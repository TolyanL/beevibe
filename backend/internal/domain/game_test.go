package domain

import (
	"errors"
	"strings"
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
		{name: "patching back to ready", from: StatusPatching, to: StatusReady},
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

func TestNewGameStartsPrivate(t *testing.T) {
	t.Parallel()

	game, err := NewGame(uuid.New(), "змейка", GameParams{})
	if err != nil {
		t.Fatalf("NewGame() error = %v, want nil", err)
	}

	if game.Visibility != VisibilityPrivate {
		t.Errorf("visibility = %q, want %q", game.Visibility, VisibilityPrivate)
	}

	if game.IsPublic() {
		t.Error("IsPublic() = true, want false for a fresh private game")
	}
}

func TestGamePublishRequiresReadyStatus(t *testing.T) {
	t.Parallel()

	game := Game{Status: StatusCoding, Visibility: VisibilityPrivate}

	if err := game.Publish(); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Publish() error = %v, want ErrInvalidTransition", err)
	}

	game.Status = StatusReady

	if err := game.Publish(); err != nil {
		t.Fatalf("Publish() error = %v, want nil", err)
	}

	if !game.IsPublic() {
		t.Error("IsPublic() = false, want true after publishing a ready game")
	}
}

func TestGameUnpublishAndDelete(t *testing.T) {
	t.Parallel()

	game := Game{Status: StatusReady, Visibility: VisibilityPublic}

	if err := game.Unpublish(); err != nil {
		t.Fatalf("Unpublish() error = %v, want nil", err)
	}

	if game.IsPublic() {
		t.Error("IsPublic() = true, want false after unpublishing")
	}

	if err := game.Delete(); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}

	if !game.IsDeleted() {
		t.Error("IsDeleted() = false, want true")
	}

	if err := game.Delete(); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Delete() error = %v, want ErrInvalidTransition for a second call", err)
	}
}

func TestGameDeleteRejectsGameBeingEdited(t *testing.T) {
	t.Parallel()

	game := Game{Status: StatusPatching}

	if err := game.Delete(); !errors.Is(err, ErrGenerationInProgress) {
		t.Fatalf("Delete() error = %v, want ErrGenerationInProgress", err)
	}
}

func TestGameRetryMovesFailedBackToQueued(t *testing.T) {
	t.Parallel()

	game := Game{Status: StatusFailed, FailReason: "llm unavailable"}

	if err := game.Retry(); err != nil {
		t.Fatalf("Retry() error = %v, want nil", err)
	}

	if game.Status != StatusQueued || game.FailReason != "" {
		t.Errorf("game = %+v, want queued status without fail reason", game)
	}

	if err := game.Retry(); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Retry() error = %v, want ErrInvalidTransition for a non-failed game", err)
	}
}

func TestGameSetMetaTrimsAndValidates(t *testing.T) {
	t.Parallel()

	game := Game{Status: StatusReady}

	cover := &GameCover{
		Palette: []string{"#000000", "#111111", "#222222", "#333333", "#ffcc00"},
		Style:   "neon",
		Seed:    "crystals",
	}

	if err := game.SetMeta("  Змейка  ", "  Собирайте линии  ", cover); err != nil {
		t.Fatalf("SetMeta() error = %v, want nil", err)
	}

	if game.Title != "Змейка" || game.Description != "Собирайте линии" || game.Cover == nil {
		t.Errorf("meta = %q/%q/%v, want trimmed values with cover", game.Title, game.Description, game.Cover)
	}

	if err := game.SetMeta(strings.Repeat("а", TitleMaxLength+1), "", nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("SetMeta() error = %v, want ErrInvalidInput for a long title", err)
	}

	if err := game.SetMeta("ok", "", &GameCover{Palette: []string{"#000000"}, Style: "neon"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("SetMeta() error = %v, want ErrInvalidInput for a short palette", err)
	}
}

func TestGameDisplayTitleFallsBackToPrompt(t *testing.T) {
	t.Parallel()

	game := Game{Prompt: "змейка в неоне"}

	if got := game.DisplayTitle(); got != "змейка в неоне" {
		t.Errorf("DisplayTitle() = %q, want the prompt", got)
	}

	game.Title = "Кристаллы"

	if got := game.DisplayTitle(); got != "Кристаллы" {
		t.Errorf("DisplayTitle() = %q, want the title", got)
	}
}

func TestNewGameSlugTransliteratesPrompt(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	slug := NewGameSlug("Неоновая змейка!", id)

	if !strings.HasPrefix(slug, "neonovaya-zmeyka-") || strings.HasSuffix(slug, "!") {
		t.Errorf("slug = %q, want transliterated prefix and id suffix", slug)
	}

	if empty := NewGameSlug("✨", id); empty != "game-1111" {
		t.Errorf("slug = %q, want game-1111 fallback", empty)
	}
}

func TestGameParamsValidateCustomLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params GameParams
		wantOK bool
	}{
		{name: "custom genre", params: GameParams{Genre: GenreCustom, CustomGenre: "ритм-игра"}, wantOK: true},
		{name: "empty custom genre", params: GameParams{Genre: GenreCustom}, wantOK: false},
		{name: "too short", params: GameParams{Genre: GenreCustom, CustomGenre: "a"}, wantOK: false},
		{name: "bad characters", params: GameParams{Genre: GenreCustom, CustomGenre: "game; drop"}, wantOK: false},
		{name: "custom style", params: GameParams{Style: StyleCustom, CustomStyle: "бумажный коллаж"}, wantOK: true},
		{name: "known style", params: GameParams{Style: StyleNeon}, wantOK: true},
	}

	for _, tt := range tests {
		err := tt.params.Validate()
		if tt.wantOK && err != nil {
			t.Errorf("%s: Validate() error = %v, want nil", tt.name, err)
		}

		if !tt.wantOK && err == nil {
			t.Errorf("%s: Validate() error = nil, want an error", tt.name)
		}
	}
}
