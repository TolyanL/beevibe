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

	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return(nil, nil).Maybe()
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{}, nil).Maybe()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
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

func TestRunPipelineFailsWithoutRefundAfterLLMReply(t *testing.T) {
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

	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return(nil, nil).Maybe()
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{}, nil).Maybe()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return(nil, nil).Maybe()
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{}, nil).Maybe()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, game.OwnerID).Return(user, nil).Maybe()

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

	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return(nil, nil).Maybe()
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{}, nil).Maybe()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return(nil, nil).Maybe()
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{}, nil).Maybe()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
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

func TestRunPipelinePausesWhenReviewFindsIssues(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, _, _ := newTestUseCase(t)

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	llmMock.EXPECT().Prompt(mock.Anything).Return("system prompt", nil)

	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: analyzerJSON}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "plan"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "code"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "assets"}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text: `{"issues":[{"code":"no_touch","message":"нет тача"}],"approved":false}`,
	}, nil).Once()

	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return(nil, nil).Maybe()
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{}, nil).Maybe()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)

	if err := uc.RunPipeline(context.Background(), game); err != nil {
		t.Fatalf("RunPipeline() error = %v, want nil: a review stop is not a failure", err)
	}

	llmMock.AssertNumberOfCalls(t, "Complete", 5)
	store.AssertNotCalled(t, "CreateVersion", mock.Anything, mock.Anything)
}

func TestRunPipelineResumesFromFailedStage(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
	}

	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return([]domain.GenEvent{
		{Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"done"}`)},
		{Type: domain.GenEventStage, Payload: []byte(`{"stage":"planning","status":"done"}`)},
		{Type: domain.GenEventStage, Payload: []byte(`{"stage":"coding","status":"error"}`)},
	}, nil)
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{
		"plan": "план из прошлой попытки",
	}, nil)
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()

	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, domain.StatusCoding, "").Return(nil)

	llmMock.EXPECT().Prompt("coder").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "code"}, nil).Once()
	llmMock.EXPECT().Prompt("assets").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: "assets"}, nil).Once()
	llmMock.EXPECT().Prompt("reviewer").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: `{"issues":[]}`}, nil).Once()
	llmMock.EXPECT().Prompt("assembler").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Once()
	llmMock.EXPECT().PromptVersion().Return("v1").Maybe()

	validator.EXPECT().Check(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	validator.EXPECT().ExtractConfig(mock.Anything, gameHTML).Return(domain.GameConfig{
		Tempo:      5,
		MapSize:    domain.MapSizeMedium,
		Difficulty: "normal",
		Style:      domain.StyleNeon,
		Palette:    []string{"#111111", "#222222"},
	}, nil).Maybe()

	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().CreateVersion(mock.Anything, mock.Anything).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.Anything).Return(nil)

	if err := uc.RunPipeline(t.Context(), game); err != nil {
		t.Fatalf("RunPipeline() error = %v, want nil", err)
	}

	llmMock.AssertNotCalled(t, "Prompt", "analyzer")
	llmMock.AssertNotCalled(t, "Prompt", "planner")
}

func TestStatusOrderHelpers(t *testing.T) {
	t.Parallel()

	if !domain.StatusCoding.Before(domain.StatusAssembling) {
		t.Error("coding must run before assembling")
	}

	if domain.StatusAssembling.Before(domain.StatusCoding) {
		t.Error("assembling must not run before coding")
	}

	if got := domain.StatusPlanning.Next(); got != domain.StatusCoding {
		t.Errorf("Planning.Next() = %s, want coding", got)
	}

	if got := domain.StatusChecking.Next(); got != domain.StatusAnalyzing {
		t.Errorf("Checking.Next() = %s, want analyzing restart", got)
	}

	for _, stage := range []domain.Status{domain.StatusCoding, domain.StatusAssembling} {
		if !stage.IsPipeline() {
			t.Errorf("%s must be a pipeline stage", stage)
		}
	}

	if domain.StatusReady.IsPipeline() || domain.StatusPatching.IsPipeline() {
		t.Error("ready and patching are not pipeline stages")
	}
}
