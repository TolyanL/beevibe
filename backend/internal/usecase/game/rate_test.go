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

func TestRateStoresRatingAndReturnsAggregate(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().UpsertRating(mock.Anything, mock.MatchedBy(func(rating domain.Rating) bool {
		return rating.GameID == gameID && rating.UserID == userID && rating.Value == domain.RatingCool
	})).Return(nil)
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{
		Cool: 3,
		Bad:  1,
	}, nil)

	output, err := uc.Rate(context.Background(), dto.RateGameInput{
		UserID: userID,
		GameID: gameID,
		Value:  "cool",
	})
	if err != nil {
		t.Fatalf("Rate() error = %v, want nil", err)
	}

	if output.Rating != "cool" {
		t.Errorf("rating = %q, want %q", output.Rating, "cool")
	}

	if output.Aggregate.Cool != 3 || output.Aggregate.Bad != 1 {
		t.Errorf("aggregate = %+v, want cool 3 bad 1", output.Aggregate)
	}
}

func TestRateRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusReady,
	}, nil)

	_, err := uc.Rate(context.Background(), dto.RateGameInput{
		UserID: uuid.New(),
		GameID: gameID,
		Value:  "cool",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Rate() error = %v, want ErrForbidden", err)
	}
}

func TestRateRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	uc, _, _, _, _ := newTestUseCase(t)

	_, err := uc.Rate(context.Background(), dto.RateGameInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Value:  "awesome",
	})
	if err == nil {
		t.Fatal("Rate() error = nil, want validation error")
	}
}

func TestRecordPlaytimeStoresSession(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().InsertPlaySession(mock.Anything, mock.MatchedBy(func(session domain.PlaySession) bool {
		return session.GameID == gameID && session.Seconds == 120 && session.UserID != nil && *session.UserID == userID
	})).Return(nil)
	store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{Seconds: 900}, nil)

	output, err := uc.RecordPlaytime(context.Background(), dto.RecordPlaytimeInput{
		UserID:  userID,
		GameID:  gameID,
		Seconds: 120,
	})
	if err != nil {
		t.Fatalf("RecordPlaytime() error = %v, want nil", err)
	}

	if output.TotalSeconds != 900 {
		t.Errorf("total seconds = %d, want 900", output.TotalSeconds)
	}
}

func TestRecordPlaytimeStoresMaxSessionLength(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().InsertPlaySession(mock.Anything, mock.MatchedBy(func(session domain.PlaySession) bool {
		return session.Seconds == domain.MaxPlaytimeSeconds
	})).Return(nil)
	store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{Seconds: 300}, nil)

	if _, err := uc.RecordPlaytime(context.Background(), dto.RecordPlaytimeInput{
		UserID:  userID,
		GameID:  gameID,
		Seconds: domain.MaxPlaytimeSeconds,
	}); err != nil {
		t.Fatalf("RecordPlaytime() error = %v, want nil", err)
	}
}

func TestRecordPlaytimeRejectsTooLongSession(t *testing.T) {
	t.Parallel()

	uc, _, _, _, _ := newTestUseCase(t)

	_, err := uc.RecordPlaytime(context.Background(), dto.RecordPlaytimeInput{
		UserID:  uuid.New(),
		GameID:  uuid.New(),
		Seconds: domain.MaxPlaytimeSeconds + 1,
	})
	if err == nil {
		t.Fatal("RecordPlaytime() error = nil, want validation error")
	}
}

func TestRollbackCreatesVersionFromSource(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	game := domain.Game{ID: gameID, OwnerID: userID, Status: domain.StatusReady}
	source := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)
	store.EXPECT().GetVersion(mock.Anything, gameID, 1).Return(source, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 3,
		HTML:   gameHTML,
	}, nil)
	store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.Source == domain.VersionSourceRollback && version.Number == 4 && version.HTML == gameHTML
	})).Return(nil)

	output, err := uc.Rollback(context.Background(), dto.RollbackGameInput{
		UserID:  userID,
		GameID:  gameID,
		Version: 1,
	})
	if err != nil {
		t.Fatalf("Rollback() error = %v, want nil", err)
	}

	if output.VersionID == uuid.Nil {
		t.Error("version id = nil, want uuid")
	}
}

func TestRollbackRejectsGameInProgress(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusPatching,
	}, nil)

	_, err := uc.Rollback(context.Background(), dto.RollbackGameInput{
		UserID:  userID,
		GameID:  gameID,
		Version: 1,
	})
	if !errors.Is(err, domain.ErrGenerationInProgress) {
		t.Fatalf("Rollback() error = %v, want ErrGenerationInProgress", err)
	}
}
