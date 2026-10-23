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

	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{}, domain.ErrVersionNotFound)

	_, err := uc.GetPublic(context.Background(), dto.GetGamePublicInput{GameID: gameID})
	if !errors.Is(err, domain.ErrVersionNotFound) {
		t.Fatalf("GetPublic() error = %v, want ErrVersionNotFound", err)
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
