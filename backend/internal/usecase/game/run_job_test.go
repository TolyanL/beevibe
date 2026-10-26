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

func TestRunNextGameJobReturnsFalseWhenQueueIsEmpty(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	store.EXPECT().ClaimGameJob(mock.Anything).Return(domain.GameJob{}, domain.ErrNotFound)

	ran, err := uc.RunNextGameJob(context.Background())
	if err != nil {
		t.Fatalf("RunNextGameJob() error = %v, want nil", err)
	}

	if ran {
		t.Error("RunNextGameJob() = true, want false on an empty queue")
	}
}

func TestRunNextGameJobClaimsAndRunsJob(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)

	ownerID := uuid.New()
	gameID := uuid.New()
	jobID := uuid.New()

	store.EXPECT().ClaimGameJob(mock.Anything).Return(domain.GameJob{
		ID:          jobID,
		GameID:      gameID,
		OwnerID:     ownerID,
		Kind:        domain.JobKindPatch,
		Instruction: "сделай фон темнее",
	}, nil)
	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: ownerID,
		Status:  domain.StatusPatching,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML,
	}, nil)
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{
		{ID: uuid.New(), GameID: gameID, Number: 1, Source: domain.VersionSourcePatch, Note: "ускорить анимацию"},
	}, nil)

	llmMock.EXPECT().Prompt("patch").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text:  gameHTML,
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 20},
	}, nil)
	llmMock.EXPECT().PromptVersion().Return("v1")
	validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil)

	store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.GameID == gameID && version.Number == 2 &&
			version.Source == domain.VersionSourcePatch && version.Note == "сделай фон темнее"
	})).Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusChecking, "").Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusReady, "").Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.MatchedBy(func(event domain.Event) bool {
		return event.Type == domain.EventGamePatched
	})).Return(nil)

	ran, err := uc.RunNextGameJob(context.Background())
	if err != nil {
		t.Fatalf("RunNextGameJob() error = %v, want nil", err)
	}

	if !ran {
		t.Error("RunNextGameJob() = false, want true when a job was claimed")
	}
}

func TestRunPatchJobKeepsChargeAndRestoresReadyAfterLLMReply(t *testing.T) {
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
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{}, nil)

	llmMock.EXPECT().Prompt("patch").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil)

	validator.EXPECT().Check(mock.Anything, gameHTML).Return([]domain.Issue{
		{Code: "external_url", Message: "external url found"},
	}, nil)

	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusReady, "").Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)

	job := domain.GameJob{ID: uuid.New(), GameID: gameID, OwnerID: ownerID, Kind: domain.JobKindPatch, Instruction: "сделай фон темнее"}

	if err := uc.RunJob(context.Background(), job); err == nil {
		t.Fatal("RunJob() error = nil, want validation failure")
	}
}

func TestRunFixJobCreatesRepairVersion(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, validator, outbox := newTestUseCase(t)

	ownerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: ownerID,
		Status:  domain.StatusPatching,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID: uuid.New(), GameID: gameID, Number: 2, HTML: gameHTML,
	}, nil)

	llmMock.EXPECT().Prompt("repair").Return("system prompt", nil).Maybe()
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil).Maybe()
	validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil).Maybe()
	llmMock.EXPECT().PromptVersion().Return("v1").Maybe()

	store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.Source == domain.VersionSourceRepair && version.Number == 3
	})).Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusChecking, "").Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusReady, "").Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)
	outbox.EXPECT().InsertOutboxEvent(mock.Anything, mock.Anything).Return(nil)

	job := domain.GameJob{ID: uuid.New(), GameID: gameID, OwnerID: ownerID, Kind: domain.JobKindFix}

	if err := uc.RunJob(context.Background(), job); err != nil {
		t.Fatalf("RunJob() error = %v, want nil", err)
	}
}

func TestRunPatchJobRefundsWhenModelNeverAnswered(t *testing.T) {
	t.Parallel()

	uc, store, llmMock, _, _ := newTestUseCase(t)

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
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{}, nil)

	llmMock.EXPECT().Prompt("patch").Return("system prompt", nil)
	llmMock.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{}, errors.New("llm down"))

	store.EXPECT().GetUserByIDForUpdate(mock.Anything, ownerID).Return(domain.User{
		ID:           ownerID,
		TokenBalance: 960,
	}, nil)
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.TokenBalance == 1000
	})).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == 40 && op.Reason == domain.ReasonRefund
	})).Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusReady, "").Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)

	job := domain.GameJob{ID: uuid.New(), GameID: gameID, OwnerID: ownerID, Kind: domain.JobKindPatch, Instruction: "сделай фон темнее"}

	if err := uc.RunJob(context.Background(), job); err == nil {
		t.Fatal("RunJob() error = nil, want llm error")
	}
}

func TestCancelStopsRunningJobAndReportsStatus(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	ownerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: ownerID,
		Status:  domain.StatusPatching,
	}, nil).Times(2)

	ctx, cancel := context.WithCancel(context.Background())

	uc.mu.Lock()
	uc.running[gameID] = &runningTask{cancel: cancel}
	uc.mu.Unlock()

	go func() {
		<-ctx.Done()

		uc.mu.Lock()
		delete(uc.running, gameID)
		uc.mu.Unlock()
	}()

	output, err := uc.Cancel(context.Background(), dto.CancelGameInput{UserID: ownerID, GameID: gameID})
	if err != nil {
		t.Fatalf("Cancel() error = %v, want nil", err)
	}

	if output.Status != string(domain.StatusPatching) {
		t.Errorf("status = %q, want patching", output.Status)
	}
}

func TestCancelRejectsIdleGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	ownerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: ownerID,
		Status:  domain.StatusReady,
	}, nil)

	_, err := uc.Cancel(context.Background(), dto.CancelGameInput{UserID: ownerID, GameID: gameID})
	if !errors.Is(err, domain.ErrNothingToCancel) {
		t.Fatalf("Cancel() error = %v, want ErrNothingToCancel", err)
	}
}

func TestCancelRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusCoding,
	}, nil)

	_, err := uc.Cancel(context.Background(), dto.CancelGameInput{UserID: uuid.New(), GameID: gameID})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Cancel() error = %v, want ErrForbidden", err)
	}
}
