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

func TestResolveReviewAcceptContinuesWithAssembly(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusNeedsDecision,
	}, nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusQueued, "").Return(nil)
	store.EXPECT().SaveGameResumeStage(mock.Anything, gameID, domain.StatusAssembling).Return(nil)

	output, err := uc.ResolveReview(context.Background(), dto.ResolveReviewInput{
		UserID: userID,
		GameID: gameID,
		Action: "accept",
	})
	if err != nil {
		t.Fatalf("ResolveReview() error = %v, want nil", err)
	}

	if output.ResumeFrom != string(domain.StatusAssembling) {
		t.Errorf("resumeFrom = %q, want assembling: the code is kept as is", output.ResumeFrom)
	}
}

func TestResolveReviewRegenerateRestartsCoding(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusNeedsDecision,
	}, nil)
	store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusQueued, "").Return(nil)
	store.EXPECT().SaveGameResumeStage(mock.Anything, gameID, domain.StatusCoding).Return(nil)

	output, err := uc.ResolveReview(context.Background(), dto.ResolveReviewInput{
		UserID: userID,
		GameID: gameID,
		Action: "regenerate",
	})
	if err != nil {
		t.Fatalf("ResolveReview() error = %v, want nil", err)
	}

	if output.ResumeFrom != string(domain.StatusCoding) {
		t.Errorf("resumeFrom = %q, want coding", output.ResumeFrom)
	}
}

func TestResolveReviewRejectsGameWithoutDecision(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)

	_, err := uc.ResolveReview(context.Background(), dto.ResolveReviewInput{
		UserID: userID,
		GameID: gameID,
		Action: "accept",
	})
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("ResolveReview() error = %v, want ErrInvalidTransition", err)
	}
}

func TestResolveReviewRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusNeedsDecision,
	}, nil)

	_, err := uc.ResolveReview(context.Background(), dto.ResolveReviewInput{
		UserID: uuid.New(),
		GameID: gameID,
		Action: "accept",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ResolveReview() error = %v, want ErrForbidden", err)
	}
}
