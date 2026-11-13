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

const fastPlan = "## Plan\n\n1. Overview: the snake eats, grows and speeds up."

const fastAnswer = "```json\n" + analyzerJSON + "\n```\n\n" + fastPlan

func TestRunPipelineFastUsesTwoCalls(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
		Params:  domain.GameParams{Speed: domain.SpeedFast},
	}

	llmMock.EXPECT().Prompt(mock.Anything).Return("system prompt", nil)
	llmMock.EXPECT().PromptVersion().Return("v1")

	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text:  fastAnswer,
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 20},
	}, nil).Once()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text:  gameHTML,
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 20},
	}, nil).Once()

	validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil)
	validator.EXPECT().ExtractConfig(mock.Anything, gameHTML).Return(domain.GameConfig{
		Tempo:   5,
		MapSize: domain.MapSizeMedium,
		Style:   domain.StyleNeon,
	}, nil)

	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, artifactHTML, gameHTML).Return(nil).Once()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, artifactCode, gameHTML).Return(nil).Once()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, artifactPlan, fastPlan).Return(nil).Once()

	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return(nil, nil).Maybe()
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{}, nil).Maybe()
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

	llmMock.AssertNumberOfCalls(t, "Complete", 2)
	llmMock.AssertCalled(t, "Prompt", "specplan")
	llmMock.AssertCalled(t, "Prompt", "onepass")
	llmMock.AssertNotCalled(t, "Prompt", "reviewer")
	llmMock.AssertNotCalled(t, "Prompt", "assembler")
}

func TestRunPipelineFastStopsAtCallBudget(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)
	uc.fast.MaxCallsPerGame = 1

	game := domain.Game{
		ID:      uuid.New(),
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusAnalyzing,
		Params:  domain.GameParams{Speed: domain.SpeedFast},
	}

	llmMock.EXPECT().Prompt(mock.Anything).Return("system prompt", nil).Maybe()
	llmMock.EXPECT().PromptVersion().Return("v1").Maybe()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: fastAnswer}, nil).Once()

	validator.EXPECT().Check(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	store.EXPECT().ListGenEvents(mock.Anything, game.ID, int64(0)).Return(nil, nil).Maybe()
	store.EXPECT().ListGameArtifacts(mock.Anything, game.ID).Return(map[string]string{}, nil).Maybe()
	store.EXPECT().SaveGameArtifact(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().UpdateGameStatus(mock.Anything, game.ID, mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, game.OwnerID).Return(domain.User{ID: game.OwnerID}, nil).Maybe()
	store.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().InsertTokenOp(mock.Anything, mock.Anything).Return(nil).Maybe()
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	err := uc.RunPipeline(context.Background(), game)
	if !errors.Is(err, domain.ErrLLMBudgetExceeded) {
		t.Fatalf("RunPipeline() error = %v, want ErrLLMBudgetExceeded", err)
	}
}

func TestResumeStageForFastPipeline(t *testing.T) {
	t.Parallel()

	fast := &pipeline{game: domain.Game{Params: domain.GameParams{Speed: domain.SpeedFast}}}
	normal := &pipeline{}

	tests := []struct {
		name   string
		speed  *pipeline
		resume domain.Status
		want   domain.Status
	}{
		{name: "analyzing stays", speed: fast, resume: domain.StatusAnalyzing, want: domain.StatusAnalyzing},
		{name: "planning folds into coding", speed: fast, resume: domain.StatusPlanning, want: domain.StatusCoding},
		{name: "coding stays", speed: fast, resume: domain.StatusCoding, want: domain.StatusCoding},
		{name: "assets folds into checking", speed: fast, resume: domain.StatusAssets, want: domain.StatusChecking},
		{name: "reviewing folds into checking", speed: fast, resume: domain.StatusReviewing, want: domain.StatusChecking},
		{name: "normal keeps the stage", speed: normal, resume: domain.StatusAssets, want: domain.StatusAssets},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resumeStageFor(tt.speed, tt.resume); got != tt.want {
				t.Fatalf("resumeStageFor() = %s, want %s", got, tt.want)
			}
		})
	}
}
