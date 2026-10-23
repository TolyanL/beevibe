package game

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/llm"
)

func TestFixRepairsBrokenGame(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	game := domain.Game{ID: gameID, OwnerID: userID, Status: domain.StatusReady}
	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 2, HTML: gameHTML}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil)

	llmMock.EXPECT().Prompt("repair").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text:  gameHTML,
		Usage: llm.Usage{PromptTokens: 5, CompletionTokens: 10},
	}, nil).Once()

	validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil)

	store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.Source == domain.VersionSourceRepair && version.Number == 3
	})).Return(nil)

	output, err := uc.Fix(context.Background(), dto.FixGameInput{
		UserID: userID,
		GameID: gameID,
		Error:  "не работает прыжок",
	})
	if err != nil {
		t.Fatalf("Fix() error = %v, want nil", err)
	}

	if output.VersionID == uuid.Nil {
		t.Error("version id = nil, want uuid")
	}
}

func TestFixRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusReady,
	}, nil)

	_, err := uc.Fix(context.Background(), dto.FixGameInput{
		UserID: uuid.New(),
		GameID: gameID,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Fix() error = %v, want ErrForbidden", err)
	}
}

func TestFixRejectsGameInProgress(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusAssembling,
	}, nil)

	_, err := uc.Fix(context.Background(), dto.FixGameInput{
		UserID: userID,
		GameID: gameID,
	})
	if !errors.Is(err, domain.ErrGenerationInProgress) {
		t.Fatalf("Fix() error = %v, want ErrGenerationInProgress", err)
	}
}

func TestFixStopsAfterFixLimit(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	game := domain.Game{ID: gameID, OwnerID: userID, Status: domain.StatusReady}
	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil)

	llmMock.EXPECT().Prompt("repair").Return("system prompt", nil).Times(2)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Times(2)

	validator.EXPECT().Check(mock.Anything, gameHTML).Return([]domain.Issue{
		{Code: "js_syntax", Message: "syntax error"},
	}, nil).Times(3)

	_, err := uc.Fix(context.Background(), dto.FixGameInput{
		UserID: userID,
		GameID: gameID,
		Error:  "игра падает",
	})
	if !errors.Is(err, domain.ErrFixLimitReached) {
		t.Fatalf("Fix() error = %v, want ErrFixLimitReached", err)
	}
}
