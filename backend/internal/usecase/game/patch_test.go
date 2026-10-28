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

func TestPatchChargesAndEnqueuesJob(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML,
	}, nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 1000,
	}, nil)
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(updated domain.User) bool {
		return updated.TokenBalance == 960
	})).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == -40 && op.Reason == domain.ReasonPatch
	})).Return(nil)
	store.EXPECT().InsertGameJob(mock.Anything, mock.MatchedBy(func(job domain.GameJob) bool {
		return job.GameID == gameID && job.OwnerID == userID &&
			job.Kind == domain.JobKindPatch && job.Instruction == "добавь второй уровень"
	})).Return(nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusPatching, "").Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.MatchedBy(func(event domain.GenEvent) bool {
		return event.Type == domain.GenEventStage && event.GameID == gameID
	})).Return(nil)

	output, err := uc.Patch(context.Background(), dto.PatchGameInput{
		UserID: userID,
		GameID: gameID,
		Text:   "добавь второй уровень",
	})
	if err != nil {
		t.Fatalf("Patch() error = %v, want nil", err)
	}

	if output.Status != string(domain.StatusPatching) {
		t.Errorf("status = %q, want %q", output.Status, domain.StatusPatching)
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
		Status:  domain.StatusPatching,
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

func TestPatchRejectsInsufficientBalance(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML,
	}, nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 10,
	}, nil)

	_, err := uc.Patch(context.Background(), dto.PatchGameInput{
		UserID: userID,
		GameID: gameID,
		Text:   "добавь второй уровень",
	})
	if !errors.Is(err, domain.ErrNotEnoughTokens) {
		t.Fatalf("Patch() error = %v, want ErrNotEnoughTokens", err)
	}
}
