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

func TestFixEnqueuesJob(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().InsertGameJob(mock.Anything, mock.MatchedBy(func(job domain.GameJob) bool {
		return job.GameID == gameID && job.Kind == domain.JobKindFix && job.Instruction == "не работает прыжок"
	})).Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusPatching, "").Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)

	output, err := uc.Fix(context.Background(), dto.FixGameInput{
		UserID: userID,
		GameID: gameID,
		Error:  "не работает прыжок",
	})
	if err != nil {
		t.Fatalf("Fix() error = %v, want nil", err)
	}

	if output.Status != string(domain.StatusPatching) {
		t.Errorf("status = %q, want %q", output.Status, domain.StatusPatching)
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

func TestRunFixJobStopsAfterFixLimit(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, _ := newTestUseCase(t)

	ownerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: ownerID,
		Status:  domain.StatusPatching,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML,
	}, nil)

	llmMock.EXPECT().Prompt("repair").Return("system prompt", nil).Times(2)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Times(2)

	validator.EXPECT().Check(mock.Anything, gameHTML).Return([]domain.Issue{
		{Code: "js_syntax", Message: "syntax error"},
	}, nil).Times(3)

	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusReady, "").Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)

	job := domain.GameJob{ID: uuid.New(), GameID: gameID, OwnerID: ownerID, Kind: domain.JobKindFix}

	err := uc.RunJob(context.Background(), job)
	if !errors.Is(err, domain.ErrFixLimitReached) {
		t.Fatalf("RunJob() error = %v, want ErrFixLimitReached", err)
	}
}
