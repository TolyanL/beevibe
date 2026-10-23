package game

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestUpdateConfigAppliesValuesToHTML(t *testing.T) {
	t.Parallel()

	uc, store, _, validator, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()
	tempo := 7

	game := domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
		Config:  &domain.GameConfig{Tempo: 5, MapSize: domain.MapSizeMedium, Style: domain.StyleNeon},
	}
	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil)

	validator.EXPECT().ApplyConfig(mock.Anything, gameHTML, mock.MatchedBy(func(cfg domain.GameConfig) bool {
		return cfg.Tempo == 7 && cfg.MapSize == domain.MapSizeMedium && cfg.Style == domain.StyleNeon
	})).Return("updated html", nil)

	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 3,
		HTML:   gameHTML,
	}, nil)
	store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.Source == domain.VersionSourceConfig && version.HTML == "updated html"
	})).Return(nil)
	store.EXPECT().UpdateGame(mock.Anything, mock.MatchedBy(func(updated domain.Game) bool {
		return updated.Config != nil && updated.Config.Tempo == 7
	})).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)

	output, err := uc.UpdateConfig(context.Background(), dto.UpdateGameConfigInput{
		UserID: userID,
		GameID: gameID,
		Tempo:  &tempo,
	})
	if err != nil {
		t.Fatalf("UpdateConfig() error = %v, want nil", err)
	}

	if output.VersionID == uuid.Nil {
		t.Error("version id = nil, want uuid")
	}
}

func TestUpdateConfigRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusReady,
	}, nil)

	_, err := uc.UpdateConfig(context.Background(), dto.UpdateGameConfigInput{
		UserID: uuid.New(),
		GameID: gameID,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("UpdateConfig() error = %v, want ErrForbidden", err)
	}
}

func TestUpdateConfigRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	uc, _, _, _, _ := newTestUseCase(t)

	tempo := 99

	_, err := uc.UpdateConfig(context.Background(), dto.UpdateGameConfigInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Tempo:  &tempo,
	})
	if err == nil {
		t.Fatal("UpdateConfig() error = nil, want validation error")
	}
}

func TestErrorCodeMapsDomainErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "llm unavailable", err: domain.ErrLLMUnavailable, want: "llm_unavailable"},
		{name: "llm budget exceeded", err: domain.ErrLLMBudgetExceeded, want: "llm_budget_exceeded"},
		{name: "validation failed", err: domain.ErrValidationFailed, want: "validation_failed"},
		{name: "not enough tokens", err: domain.ErrNotEnoughTokens, want: "not_enough_tokens"},
		{name: "timeout", err: context.DeadlineExceeded, want: "timeout"},
		{name: "unknown", err: errors.New("boom"), want: "generation_failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := errorCode(tt.err); got != tt.want {
				t.Errorf("errorCode() = %q, want %q", got, tt.want)
			}
		})
	}
}
