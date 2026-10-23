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

func TestPatchCreatesNewVersion(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	game := domain.Game{ID: gameID, OwnerID: userID, Status: domain.StatusReady}
	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil)
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{latest}, nil)

	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 1000,
	}, nil)
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(updated domain.User) bool {
		return updated.TokenBalance == 1000-40
	})).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == -40 && op.Reason == domain.ReasonPatch
	})).Return(nil)

	llmMock.EXPECT().Prompt("patch").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text:  gameHTML,
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 20},
	}, nil).Once()
	llmMock.EXPECT().PromptVersion().Return("v1")

	validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil)

	store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.GameID == gameID && version.Number == 2 && version.Source == domain.VersionSourcePatch
	})).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.MatchedBy(func(event domain.Event) bool {
		return event.Type == domain.EventGamePatched
	})).Return(nil)

	output, err := uc.Patch(context.Background(), dto.PatchGameInput{
		UserID: userID,
		GameID: gameID,
		Text:   "добавь второй уровень",
	})
	if err != nil {
		t.Fatalf("Patch() error = %v, want nil", err)
	}

	if output.VersionID == uuid.Nil {
		t.Error("version id = nil, want uuid")
	}

	if output.TokensSpent != 30 {
		t.Errorf("tokens spent = %d, want 30", output.TokensSpent)
	}
}

func TestPatchRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusReady,
	}, nil)

	_, err := uc.Patch(context.Background(), dto.PatchGameInput{
		UserID: uuid.New(),
		GameID: gameID,
		Text:   "добавь второй уровень",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Patch() error = %v, want ErrForbidden", err)
	}
}

func TestPatchRejectsGameInProgress(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusCoding,
	}, nil)

	_, err := uc.Patch(context.Background(), dto.PatchGameInput{
		UserID: userID,
		GameID: gameID,
		Text:   "добавь второй уровень",
	})
	if !errors.Is(err, domain.ErrGenerationInProgress) {
		t.Fatalf("Patch() error = %v, want ErrGenerationInProgress", err)
	}
}

func TestPatchRefundsWhenValidationFails(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	game := domain.Game{ID: gameID, OwnerID: userID, Status: domain.StatusReady}
	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil)
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{latest}, nil)

	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 1000,
	}, nil).Once()
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(updated domain.User) bool {
		return updated.TokenBalance == 960
	})).Return(nil).Once()
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == -40
	})).Return(nil).Once()

	llmMock.EXPECT().Prompt("patch").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Once()

	validator.EXPECT().Check(mock.Anything, gameHTML).Return([]domain.Issue{
		{Code: "external_url", Message: "external url found"},
	}, nil)

	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 960,
	}, nil).Once()
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(updated domain.User) bool {
		return updated.TokenBalance == 1000
	})).Return(nil).Once()
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == 40 && op.Reason == domain.ReasonRefund
	})).Return(nil).Once()

	_, err := uc.Patch(context.Background(), dto.PatchGameInput{
		UserID: userID,
		GameID: gameID,
		Text:   "добавь второй уровень",
	})
	if !errors.Is(err, domain.ErrValidationFailed) {
		t.Fatalf("Patch() error = %v, want ErrValidationFailed", err)
	}
}
