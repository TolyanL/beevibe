package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

const analyzerJSON = `{"genre":"snake","mechanic":"classic snake","controls":"arrows and swipe",` +
	`"win_condition":"fill the board","lose_condition":"hit the wall","palette":["#101010","#33ff66"],` +
	`"map_size":"medium","style":"neon","audience":"grandma","assets":["apple","snake"]}`

const reviewerJSON = `{"issues":[],"approved":true}`

const gameHTML = `<!DOCTYPE html><html><body><canvas id="c"></canvas><script>const CONFIG = {tempo: 5};</script></body></html>`

func TestRunPipelineReachesReady(t *testing.T) {
	t.Parallel()

	uc, store, llm, validator, _ := newTestUseCase(t)

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	llm.EXPECT().Prompt(mock.Anything).Return("system prompt", nil)
	llm.EXPECT().PromptVersion().Return("v1")

	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{
		Text:  analyzerJSON,
		Usage: dto.LLMUsage{PromptTokens: 10, CompletionTokens: 20},
	}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: "plan"}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: "code"}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: "assets"}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: reviewerJSON}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: gameHTML}, nil).Once()

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

	if err := uc.RunPipeline(context.Background(), game); err != nil {
		t.Fatalf("RunPipeline() error = %v, want nil", err)
	}
}

func TestRunPipelineFailsAndRefundsWhenValidationFails(t *testing.T) {
	t.Parallel()

	uc, store, llm, validator, _ := newTestUseCase(t)

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	user := domain.User{ID: game.OwnerID, TokenBalance: 0}

	llm.EXPECT().Prompt(mock.Anything).Return("system prompt", nil).Maybe()
	llm.EXPECT().PromptVersion().Return("v1").Maybe()

	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: analyzerJSON}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: "plan"}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: "code"}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: "assets"}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: reviewerJSON}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: gameHTML}, nil).Once()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: gameHTML}, nil).Maybe()

	validator.EXPECT().Check(mock.Anything, mock.Anything).Return([]domain.Issue{
		{Code: "external_url", Message: "external url found"},
	}, nil)

	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
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

	uc, store, llm, _, _ := newTestUseCase(t)
	uc.worker.MaxCallsPerGame = 1

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	user := domain.User{ID: game.OwnerID, TokenBalance: 0}

	llm.EXPECT().Prompt(mock.Anything).Return("system prompt", nil).Maybe()
	llm.EXPECT().PromptVersion().Return("v1").Maybe()
	llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(dto.LLMResponse{Text: analyzerJSON}, nil).Once()

	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, game.OwnerID).Return(user, nil).Maybe()
	store.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().InsertTokenOp(mock.Anything, mock.Anything).Return(nil).Maybe()

	err := uc.RunPipeline(context.Background(), game)
	if !errors.Is(err, domain.ErrLLMBudgetExceeded) {
		t.Fatalf("RunPipeline() error = %v, want ErrLLMBudgetExceeded", err)
	}
}
