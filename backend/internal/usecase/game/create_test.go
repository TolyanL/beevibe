package game

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	gamemocks "github.com/TolyanL/beevibe/internal/usecase/mocks/game"
	llmmocks "github.com/TolyanL/beevibe/internal/usecase/mocks/llm"
)

func newTestUseCase(t *testing.T) (*UseCase, *gamemocks.MockStore, *llmmocks.MockLLM, *gamemocks.MockValidator, *gamemocks.MockOutbox) {
	t.Helper()

	store := gamemocks.NewMockStore(t)
	llmMock := llmmocks.NewMockLLM(t)
	validator := gamemocks.NewMockValidator(t)
	outbox := gamemocks.NewMockOutbox(t)

	uc := New(store, outbox, llmMock, validator,
		Tokens{
			PriceGenerate: 320,
			PricePatch:    40,
		},
		Limits{
			MaxTokensAnalyze:  100,
			MaxTokensPlan:     100,
			MaxTokensCode:     100,
			MaxTokensAssets:   100,
			MaxTokensReview:   100,
			MaxTokensAssemble: 100,
			MaxTokensPatch:    100,
			MaxTokensRepair:   100,
		},
		Worker{
			Timeout:         60,
			MaxCallsPerGame: 12,
			MaxFixAttempts:  2,
		},
	)

	return uc, store, llmMock, validator, outbox
}

func TestCreateDebitsTokensAndCreatesGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, outbox := newTestUseCase(t)

	userID := uuid.New()
	user := domain.User{ID: userID, Email: "user@example.com", TokenBalance: 1000}

	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(user, nil)
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(updated domain.User) bool {
		return updated.TokenBalance == 1000-320
	})).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == -320 && op.Reason == domain.ReasonGenerate
	})).Return(nil)
	store.EXPECT().CreateGame(mock.Anything, mock.MatchedBy(func(game domain.Game) bool {
		return game.OwnerID == userID && game.Status == domain.StatusQueued && game.Prompt == "змейка"
	})).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.MatchedBy(func(event domain.Event) bool {
		return event.Type == domain.EventGameCreated && event.Topic == domain.GameEventsTopic
	})).Return(nil)

	output, err := uc.Create(context.Background(), dto.CreateGameInput{
		UserID: userID,
		Prompt: "змейка",
	})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	if output.ID == uuid.Nil {
		t.Fatal("Create() id = nil, want uuid")
	}

	if output.Status != string(domain.StatusQueued) {
		t.Fatalf("status = %s, want %s", output.Status, domain.StatusQueued)
	}
}

func TestCreateFailsWhenNotEnoughTokens(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	user := domain.User{ID: userID, Email: "user@example.com", TokenBalance: 10}

	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(user, nil)

	_, err := uc.Create(context.Background(), dto.CreateGameInput{
		UserID: userID,
		Prompt: "змейка",
	})
	if !errors.Is(err, domain.ErrNotEnoughTokens) {
		t.Fatalf("Create() error = %v, want ErrNotEnoughTokens", err)
	}
}

func TestCreateReplaysIdempotencyKey(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()
	game := domain.Game{ID: gameID, OwnerID: userID, Status: domain.StatusCoding}

	store.EXPECT().InsertIdempotencyKey(mock.Anything, "key-1", operationCreateGame, userID).Return(false, nil)
	store.EXPECT().GetIdempotencyResult(mock.Anything, "key-1", operationCreateGame).Return(gameID.String(), true, nil)
	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)

	output, err := uc.Create(context.Background(), dto.CreateGameInput{
		UserID:         userID,
		Prompt:         "змейка",
		IdempotencyKey: "key-1",
	})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	if output.ID != gameID {
		t.Fatalf("id = %s, want %s", output.ID, gameID)
	}
}

func TestCreateReturnsConflictWhileIdempotencyInProgress(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().InsertIdempotencyKey(mock.Anything, "key-2", operationCreateGame, userID).Return(false, nil)
	store.EXPECT().GetIdempotencyResult(mock.Anything, "key-2", operationCreateGame).Return("", true, nil)

	_, err := uc.Create(context.Background(), dto.CreateGameInput{
		UserID:         userID,
		Prompt:         "змейка",
		IdempotencyKey: "key-2",
	})
	if !errors.Is(err, domain.ErrIdempotencyInProgress) {
		t.Fatalf("Create() error = %v, want ErrIdempotencyInProgress", err)
	}
}
