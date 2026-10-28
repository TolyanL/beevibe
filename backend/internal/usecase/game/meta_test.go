package game

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestUpdateMetaSavesTrimmedTitle(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().UpdateGameMeta(mock.Anything, mock.MatchedBy(func(game domain.Game) bool {
		return game.Title == "Кристаллы" && game.Description == "Собирайте линии"
	})).Return(nil)

	title := "  Кристаллы  "
	description := "Собирайте линии"

	output, err := uc.UpdateMeta(context.Background(), dto.UpdateGameMetaInput{
		UserID:      userID,
		GameID:      gameID,
		Title:       &title,
		Description: &description,
	})
	if err != nil {
		t.Fatalf("UpdateMeta() error = %v, want nil", err)
	}

	if output.Title != "Кристаллы" || output.Description != "Собирайте линии" {
		t.Errorf("output = %+v, want trimmed meta", output)
	}
}

func TestUpdateMetaRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
	}, nil)

	title := "чужое"

	_, err := uc.UpdateMeta(context.Background(), dto.UpdateGameMetaInput{
		UserID: uuid.New(),
		GameID: gameID,
		Title:  &title,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("UpdateMeta() error = %v, want ErrForbidden", err)
	}
}

func TestGenerateMetaChargesAndEnqueuesJob(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 50_000,
	}, nil)
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.TokenBalance == 50_000-10
	})).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == -10 && op.Reason == domain.ReasonMeta
	})).Return(nil)
	store.EXPECT().InsertGameJob(mock.Anything, mock.MatchedBy(func(job domain.GameJob) bool {
		return job.Kind == domain.JobKindMeta && job.GameID == gameID
	})).Return(nil)
	store.EXPECT().InsertGenEvent(mock.Anything, mock.MatchedBy(func(event domain.GenEvent) bool {
		return event.GameID == gameID && event.Type == domain.GenEventStage
	})).Return(nil)

	output, err := uc.GenerateMeta(context.Background(), dto.GenerateGameMetaInput{UserID: userID, GameID: gameID})
	if err != nil {
		t.Fatalf("GenerateMeta() error = %v, want nil", err)
	}

	if output.Status != "queued" {
		t.Errorf("status = %q, want queued", output.Status)
	}
}

func TestGenerateMetaRejectsGameInProgress(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusPatching,
	}, nil)

	_, err := uc.GenerateMeta(context.Background(), dto.GenerateGameMetaInput{UserID: userID, GameID: gameID})
	if !errors.Is(err, domain.ErrGenerationInProgress) {
		t.Fatalf("GenerateMeta() error = %v, want ErrGenerationInProgress", err)
	}
}

func TestGetReportsMetaOperationWhileGenerating(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()
	startedAt := time.Now().UTC()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{}, nil)
	store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return([]domain.GenEvent{
		{
			GameID:    gameID,
			Seq:       7,
			Type:      domain.GenEventStage,
			Payload:   []byte(`{"stage":"meta","status":"running","operation":"meta"}`),
			CreatedAt: startedAt,
		},
	}, nil)
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, userID).Return(domain.Rating{}, domain.ErrNotFound)

	output, err := uc.Get(context.Background(), dto.GetGameInput{UserID: userID, GameID: gameID})
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if output.Operation == nil || output.Operation.Kind != string(domain.JobKindMeta) {
		t.Fatalf("operation = %+v, want meta operation while the title is generated", output.Operation)
	}

	if !output.Operation.StartedAt.Equal(startedAt) {
		t.Errorf("startedAt = %v, want %v", output.Operation.StartedAt, startedAt)
	}
}

func TestGetHidesMetaOperationAfterDone(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{}, nil)
	store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return([]domain.GenEvent{
		{
			GameID:  gameID,
			Seq:     8,
			Type:    domain.GenEventStage,
			Payload: []byte(`{"stage":"meta","status":"done","operation":"meta"}`),
		},
	}, nil)
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, userID).Return(domain.Rating{}, domain.ErrNotFound)

	output, err := uc.Get(context.Background(), dto.GetGameInput{UserID: userID, GameID: gameID})
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if output.Operation != nil {
		t.Errorf("operation = %+v, want nil after the meta job finished", output.Operation)
	}
}
