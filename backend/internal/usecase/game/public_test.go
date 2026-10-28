package game

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestGetPublicCardReturnsCardWithoutViewer(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()
	createdAt := time.Now().UTC()

	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:        gameID,
		OwnerID:   uuid.New(),
		Prompt:    "змейка в неоне",
		Params:    domain.GameParams{Genre: domain.GenreSnake, Style: domain.StyleNeon, MapSize: domain.MapSizeMedium},
		Status:    domain.StatusReady,
		CreatedAt: createdAt,
	}, nil)
	store.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(domain.User{
		ID:          uuid.New(),
		Email:       "author@example.com",
		DisplayName: "Автор",
	}, nil).Maybe()
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{Average: 4.5, Total: 2}, nil)
	store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{Sessions: 3, Seconds: 5400, Hours: 1.5}, nil)

	output, err := uc.GetPublicCard(context.Background(), dto.GetGamePublicCardInput{Slug: "zmeyka-1a2b"})
	if err != nil {
		t.Fatalf("GetPublicCard() error = %v, want nil", err)
	}

	if output.ID != gameID || output.Prompt != "змейка в неоне" {
		t.Errorf("card = %+v, want card of game %s", output, gameID)
	}

	if output.Genre != string(domain.GenreSnake) || output.Style != string(domain.StyleNeon) || output.MapSize != string(domain.MapSizeMedium) {
		t.Errorf("params = %s/%s/%s, want snake/neon/medium", output.Genre, output.Style, output.MapSize)
	}

	if !output.CreatedAt.Equal(createdAt) {
		t.Errorf("createdAt = %v, want %v", output.CreatedAt, createdAt)
	}

	if output.MyRating != 0 {
		t.Errorf("myRating = %d, want 0 for anonymous viewer", output.MyRating)
	}
}

func TestGetPublicCardIncludesViewerRating(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	viewerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(domain.User{
		ID:          uuid.New(),
		Email:       "author@example.com",
		DisplayName: "Автор",
	}, nil).Maybe()
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, viewerID).Return(domain.Rating{Value: domain.Rating3}, nil)

	output, err := uc.GetPublicCard(context.Background(), dto.GetGamePublicCardInput{Slug: "zmeyka-1a2b", ViewerID: &viewerID})
	if err != nil {
		t.Fatalf("GetPublicCard() error = %v, want nil", err)
	}

	if output.MyRating != 3 {
		t.Errorf("myRating = %d, want 3", output.MyRating)
	}

	if output.IsOwner {
		t.Error("isOwner = true, want false for a non-owner viewer")
	}
}

func TestGetPublicCardMarksOwner(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	ownerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:      gameID,
		OwnerID: ownerID,
		Prompt:  "змейка",
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(domain.User{
		ID:          uuid.New(),
		Email:       "author@example.com",
		DisplayName: "Автор",
	}, nil).Maybe()
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, ownerID).Return(domain.Rating{}, domain.ErrNotFound)

	output, err := uc.GetPublicCard(context.Background(), dto.GetGamePublicCardInput{Slug: "zmeyka-1a2b", ViewerID: &ownerID})
	if err != nil {
		t.Fatalf("GetPublicCard() error = %v, want nil", err)
	}

	if !output.IsOwner {
		t.Error("isOwner = false, want true for the owner")
	}
}

func TestGetPublicCardToleratesMissingViewerRating(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	viewerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusReady,
	}, nil)
	store.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(domain.User{
		ID:          uuid.New(),
		Email:       "author@example.com",
		DisplayName: "Автор",
	}, nil).Maybe()
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, viewerID).Return(domain.Rating{}, domain.ErrNotFound)

	output, err := uc.GetPublicCard(context.Background(), dto.GetGamePublicCardInput{Slug: "zmeyka-1a2b", ViewerID: &viewerID})
	if err != nil {
		t.Fatalf("GetPublicCard() error = %v, want nil", err)
	}

	if output.MyRating != 0 {
		t.Errorf("myRating = %d, want 0", output.MyRating)
	}
}
