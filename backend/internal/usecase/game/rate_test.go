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
		ID:         gameID,
		OwnerID:    uuid.New(),
		Visibility: domain.VisibilityPublic,
		Status:     domain.StatusReady,
	}, nil)
	store.EXPECT().UpsertRating(mock.Anything, mock.MatchedBy(func(rating domain.Rating) bool {
		return rating.GameID == gameID && rating.UserID == userID && rating.Value == domain.Rating5
	})).Return(nil)
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{
		Average: 4.5,
		Total:   4,
		Counts:  [5]int{0, 0, 0, 2, 2},
	}, nil)

	output, err := uc.Rate(context.Background(), dto.RateGameInput{
		UserID: userID,
		GameID: gameID,
		Value:  5,
	})
	if err != nil {
		t.Fatalf("Rate() error = %v, want nil", err)
	}

	if output.Rating != 5 {
		t.Errorf("rating = %d, want %d", output.Rating, 5)
	}

	if output.Aggregate.Average != 4.5 || output.Aggregate.Total != 4 {
		t.Errorf("aggregate = %+v, want average 4.5 and 4 votes", output.Aggregate)
	}
}

func TestRateRejectsOwnerRating(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	ownerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: ownerID,
		Status:  domain.StatusReady,
	}, nil)

	_, err := uc.Rate(context.Background(), dto.RateGameInput{
		UserID: ownerID,
		GameID: gameID,
		Value:  5,
	})
	if !errors.Is(err, domain.ErrCannotRateOwnGame) {
		t.Fatalf("Rate() error = %v, want ErrCannotRateOwnGame", err)
	}
}

func TestRateRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	uc, _, _, _, _ := newTestUseCase(t)

	_, err := uc.Rate(context.Background(), dto.RateGameInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Value:  6,
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

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Visibility: domain.VisibilityPublic,
		Status:     domain.StatusReady,
	}, nil)
	store.EXPECT().InsertPlaySession(mock.Anything, mock.MatchedBy(func(session domain.PlaySession) bool {
		return session.GameID == gameID && session.Seconds == 120 && session.UserID != nil &&
			*session.UserID == userID && !session.CreatedAt.IsZero()
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

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Visibility: domain.VisibilityPublic,
		Status:     domain.StatusReady,
	}, nil)
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

func TestRateHidesUnpublishedGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Visibility: domain.VisibilityPrivate,
		Status:     domain.StatusReady,
	}, nil)

	if _, err := uc.Rate(context.Background(), dto.RateGameInput{
		UserID: uuid.New(),
		GameID: gameID,
		Value:  5,
	}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Rate() error = %v, want ErrNotFound", err)
	}
}

func TestRateHidesGameThatIsNotReady(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Visibility: domain.VisibilityPublic,
		Status:     domain.StatusCoding,
	}, nil)

	if _, err := uc.Rate(context.Background(), dto.RateGameInput{
		UserID: uuid.New(),
		GameID: gameID,
		Value:  5,
	}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Rate() error = %v, want ErrNotFound", err)
	}
}

func TestRecordPlaytimeHidesUnpublishedGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Visibility: domain.VisibilityPrivate,
		Status:     domain.StatusReady,
	}, nil)

	if _, err := uc.RecordPlaytime(context.Background(), dto.RecordPlaytimeInput{
		UserID:  uuid.New(),
		GameID:  gameID,
		Seconds: 60,
	}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("RecordPlaytime() error = %v, want ErrNotFound", err)
	}
}

func TestRecordPlaytimeLetsOwnerTrackDraft(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	ownerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    ownerID,
		Visibility: domain.VisibilityPrivate,
		Status:     domain.StatusCoding,
	}, nil)
	store.EXPECT().InsertPlaySession(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{Seconds: 60}, nil)

	if _, err := uc.RecordPlaytime(context.Background(), dto.RecordPlaytimeInput{
		UserID:  ownerID,
		GameID:  gameID,
		Seconds: 60,
	}); err != nil {
		t.Fatalf("RecordPlaytime() error = %v, want nil", err)
	}
}
