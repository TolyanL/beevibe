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

func TestGetReturnsGameWithVersionsAndCost(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	game := domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
		Spec:    &domain.GameSpec{Genre: "snake"},
		Config:  &domain.GameConfig{Tempo: 5},
	}

	versions := []domain.Version{
		{ID: uuid.New(), Number: 1, Source: domain.VersionSourceGenerate, TokensCost: 100, ModelCalls: 6},
		{ID: uuid.New(), Number: 2, Source: domain.VersionSourceRepair, TokensCost: 50, ModelCalls: 1},
	}

	events := []domain.GenEvent{
		{Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"done"}`)},
		{Type: domain.GenEventArtifact, Payload: []byte(`{"kind":"spec"}`)},
	}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)
	store.EXPECT().ListVersions(mock.Anything, gameID).Return(versions, nil)
	store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return(events, nil)
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{
		Average: 4.5,
		Total:   4,
		Counts:  [5]int{0, 0, 0, 2, 2},
	}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, userID).Return(domain.Rating{
		GameID: gameID,
		UserID: userID,
		Value:  domain.Rating5,
	}, nil)

	output, err := uc.Get(context.Background(), dto.GetGameInput{UserID: userID, GameID: gameID})
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if output.Status != string(domain.StatusReady) {
		t.Errorf("status = %q, want %q", output.Status, domain.StatusReady)
	}

	if len(output.Versions) != 2 {
		t.Fatalf("versions = %d, want 2", len(output.Versions))
	}

	if output.Cost.Tokens != 150 || output.Cost.Calls != 7 || output.Cost.Fixes != 1 {
		t.Errorf("cost = %+v, want tokens 150 calls 7 fixes 1", output.Cost)
	}

	if len(output.Stages) != 1 || output.Stages[0].Stage != "analyzing" {
		t.Errorf("stages = %+v, want one analyzing stage", output.Stages)
	}

	if output.Rating.Average != 4.5 || output.MyRating != 5 {
		t.Errorf("rating = %+v my = %d, want average 4.5 and 5", output.Rating, output.MyRating)
	}
}

func TestGetRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusReady,
	}, nil)

	_, err := uc.Get(context.Background(), dto.GetGameInput{UserID: uuid.New(), GameID: gameID})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Get() error = %v, want ErrForbidden", err)
	}
}

func TestGetPublicReturnsLatestHTML(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPublic,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 2,
		HTML:   gameHTML,
	}, nil)

	output, err := uc.GetPublic(context.Background(), dto.GetGamePublicInput{GameID: gameID})
	if err != nil {
		t.Fatalf("GetPublic() error = %v, want nil", err)
	}

	if output.HTML != gameHTML {
		t.Errorf("html = %q, want %q", output.HTML, gameHTML)
	}
}

func TestGetPublicFailsWhenVersionMissing(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPublic,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{}, domain.ErrVersionNotFound)

	_, err := uc.GetPublic(context.Background(), dto.GetGamePublicInput{GameID: gameID})
	if !errors.Is(err, domain.ErrVersionNotFound) {
		t.Fatalf("GetPublic() error = %v, want ErrVersionNotFound", err)
	}
}

func TestGetExposesPromptFailReasonAndCreatedForFailedGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()
	createdAt := time.Now().UTC()

	game := domain.Game{
		ID:         gameID,
		OwnerID:    userID,
		Prompt:     "змейка в неоновом лабиринте",
		Params:     domain.GameParams{Genre: domain.GenreSnake, MapSize: domain.MapSizeMedium, Style: domain.StyleNeon},
		Status:     domain.StatusFailed,
		FailReason: "analyzer: response is empty",
		CreatedAt:  createdAt,
	}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(game, nil)
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{}, nil)
	store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return([]domain.GenEvent{}, nil)
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, userID).Return(domain.Rating{}, domain.ErrNotFound)

	output, err := uc.Get(context.Background(), dto.GetGameInput{UserID: userID, GameID: gameID})
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if output.ID != gameID || output.Prompt != game.Prompt {
		t.Errorf("id/prompt = %s/%q, want %s/%q", output.ID, output.Prompt, gameID, game.Prompt)
	}

	if output.Params.Genre != domain.GenreSnake || output.Params.MapSize != domain.MapSizeMedium || output.Params.Style != domain.StyleNeon {
		t.Errorf("params = %+v, want snake medium neon", output.Params)
	}

	if output.FailReason != game.FailReason {
		t.Errorf("failReason = %q, want %q", output.FailReason, game.FailReason)
	}

	if !output.CreatedAt.Equal(createdAt) {
		t.Errorf("createdAt = %v, want %v", output.CreatedAt, createdAt)
	}
}

func TestListGenEventsReturnsEventsOfOwner(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	events := []domain.GenEvent{
		{Seq: 1, GameID: gameID, Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing"}`)},
	}

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusCoding,
	}, nil)
	store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return(events, nil)

	output, err := uc.ListGenEvents(context.Background(), dto.ListGenEventsInput{
		UserID: userID,
		GameID: gameID,
	})
	if err != nil {
		t.Fatalf("ListGenEvents() error = %v, want nil", err)
	}

	if len(output.Events) != 1 {
		t.Fatalf("events = %d, want 1", len(output.Events))
	}

	if output.Status != string(domain.StatusCoding) {
		t.Errorf("status = %q, want %q", output.Status, domain.StatusCoding)
	}
}

func TestListGenEventsRejectsForeignGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusCoding,
	}, nil)

	_, err := uc.ListGenEvents(context.Background(), dto.ListGenEventsInput{
		UserID: uuid.New(),
		GameID: gameID,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ListGenEvents() error = %v, want ErrForbidden", err)
	}
}

func TestGetReportsActiveOperationWhilePatching(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()
	startedAt := time.Now().UTC()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusPatching,
	}, nil)
	store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{}, nil)
	store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return([]domain.GenEvent{
		{
			GameID:    gameID,
			Type:      domain.GenEventStage,
			Payload:   []byte(`{"stage":"patching","status":"running","operation":"fix"}`),
			CreatedAt: startedAt,
		},
	}, nil)
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, userID).Return(domain.Rating{}, domain.ErrNotFound)

	output, err := uc.Get(context.Background(), dto.GetGameInput{UserID: userID, GameID: gameID})
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if output.Operation == nil {
		t.Fatal("operation = nil, want active operation while patching")
	}

	if output.Operation.Kind != string(domain.JobKindFix) {
		t.Errorf("kind = %q, want %q", output.Operation.Kind, domain.JobKindFix)
	}

	if !output.Operation.StartedAt.Equal(startedAt) {
		t.Errorf("startedAt = %v, want %v", output.Operation.StartedAt, startedAt)
	}
}

func TestGetHidesOperationWhenReady(t *testing.T) {
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
		{GameID: gameID, Type: domain.GenEventStage, Payload: []byte(`{"stage":"patching","status":"done","operation":"patch"}`)},
	}, nil)
	store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	store.EXPECT().GetRating(mock.Anything, gameID, userID).Return(domain.Rating{}, domain.ErrNotFound)

	output, err := uc.Get(context.Background(), dto.GetGameInput{UserID: userID, GameID: gameID})
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if output.Operation != nil {
		t.Errorf("operation = %+v, want nil for a ready game", output.Operation)
	}
}

func TestGetPublicHidesPrivateGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPrivate,
	}, nil)

	_, err := uc.GetPublic(context.Background(), dto.GetGamePublicInput{GameID: gameID})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetPublic() error = %v, want ErrNotFound", err)
	}
}

func TestGetPublicHidesDeletedGame(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()
	deletedAt := time.Now().UTC()

	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPublic,
		DeletedAt:  &deletedAt,
	}, nil)

	_, err := uc.GetPublic(context.Background(), dto.GetGamePublicInput{GameID: gameID})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetPublic() error = %v, want ErrNotFound", err)
	}
}
