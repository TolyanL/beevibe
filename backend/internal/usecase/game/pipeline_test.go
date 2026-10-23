package game

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/llm"
)

const analyzerJSON = `{"genre":"snake","mechanic":"classic snake","controls":"arrows and swipe",` +
	`"win_condition":"fill the board","lose_condition":"hit the wall","palette":["#101010","#33ff66"],` +
	`"map_size":"medium","style":"neon","audience":"grandma","assets":["apple","snake"]}`

const reviewerJSON = `{"issues":[],"approved":true}`

const gameHTML = `<!DOCTYPE html><html><body><canvas id="c"></canvas><script>const CONFIG = {tempo: 5};</script></body></html>`

func TestRunPipelineReachesReady(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	llmMock.EXPECT().Prompt(mock.Anything).Return("system prompt", nil)
	llmMock.EXPECT().PromptVersion().Return("v1")

	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text:  analyzerJSON,
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 20},
	}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "plan"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "code"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "assets"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: reviewerJSON}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Once()

	validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil)
	validator.EXPECT().ExtractConfig(mock.Anything, gameHTML).Return(domain.GameConfig{
		Tempo:   5,
		MapSize: domain.MapSizeMedium,
		Style:   domain.StyleNeon,
	}, nil)

	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.GameID == game.ID && version.Number == 1 && version.Source == domain.VersionSourceGenerate
	})).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.MatchedBy(func(event domain.Event) bool {
		return event.Type == domain.EventGameReady
	})).Return(nil)

	if err := uc.RunPipeline(context.Background(), game); err != nil {
		t.Fatalf("RunPipeline() error = %v, want nil", err)
	}
}

func TestRunPipelineFailsAndRefundsWhenValidationFails(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	user := domain.User{ID: game.OwnerID, TokenBalance: 0}

	llmMock.EXPECT().Prompt(mock.Anything).Return("system prompt", nil).Maybe()
	llmMock.EXPECT().PromptVersion().Return("v1").Maybe()

	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: analyzerJSON}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "plan"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "code"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "assets"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: reviewerJSON}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Maybe()

	validator.EXPECT().Check(mock.Anything, mock.Anything).Return([]domain.Issue{
		{Code: "external_url", Message: "external url found"},
	}, nil)

	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, game.OwnerID).Return(user, nil)
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(updated domain.User) bool {
		return updated.TokenBalance == 320
	})).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == 320 && op.Reason == domain.ReasonRefund
	})).Return(nil)

	err := uc.RunPipeline(context.Background(), game)
	if !errors.Is(err, domain.ErrValidationFailed) {
		t.Fatalf("RunPipeline() error = %v, want ErrValidationFailed", err)
	}
}

func TestRunPipelineStopsAtCallBudget(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, _, outbox := newTestUseCase(t)
	uc.worker.MaxCallsPerGame = 1

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	user := domain.User{ID: game.OwnerID, TokenBalance: 0}

	llmMock.EXPECT().Prompt(mock.Anything).Return("system prompt", nil).Maybe()
	llmMock.EXPECT().PromptVersion().Return("v1").Maybe()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: analyzerJSON}, nil).Once()

	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, game.OwnerID).Return(user, nil).Maybe()
	store.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().InsertTokenOp(mock.Anything, mock.Anything).Return(nil).Maybe()

	err := uc.RunPipeline(context.Background(), game)
	if !errors.Is(err, domain.ErrLLMBudgetExceeded) {
		t.Fatalf("RunPipeline() error = %v, want ErrLLMBudgetExceeded", err)
	}
}

func TestRunPipelineRetriesCodeAfterReviewIssues(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	llmMock.EXPECT().Prompt(mock.Anything).Return("system prompt", nil)
	llmMock.EXPECT().PromptVersion().Return("v1")

	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: analyzerJSON}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "plan"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "code"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "assets"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text: `{"issues":[{"code":"no_touch","message":"нет тача"}],"approved":false}`,
	}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "code v2"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "assets v2"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: reviewerJSON}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Once()

	validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil)
	validator.EXPECT().ExtractConfig(mock.Anything, gameHTML).Return(domain.GameConfig{
		Tempo:   5,
		MapSize: domain.MapSizeMedium,
		Style:   domain.StyleNeon,
	}, nil)

	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.GameID == game.ID && version.Number == 1
	})).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.Anything).Return(nil)

	if err := uc.RunPipeline(context.Background(), game); err != nil {
		t.Fatalf("RunPipeline() error = %v, want nil", err)
	}
}
