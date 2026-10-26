package game

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/dto"
)

func TestListUserGamesAppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	games := []dto.ProfileGame{{ID: uuid.New(), Prompt: "змейка", Status: "ready"}}

	store.EXPECT().ListUserGames(mock.Anything, userID, dto.ListUserGamesInput{
		Status: "ready",
		Sort:   "tokens_cost",
		Limit:  20,
		Offset: 0,
	}).Return(games, 1, nil)

	output, err := uc.ListUserGames(context.Background(), dto.ListUserGamesInput{
		UserID: userID,
		Status: "ready",
		Sort:   "tokens_cost",
	})
	if err != nil {
		t.Fatalf("ListUserGames() error = %v, want nil", err)
	}

	if len(output.Items) != 1 || output.Total != 1 {
		t.Errorf("output = %+v, want one game with total 1", output)
	}
}

func TestListUserGamesRejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	uc, _, _, _, _ := newTestUseCase(t)

	if _, err := uc.ListUserGames(context.Background(), dto.ListUserGamesInput{
		UserID: uuid.New(),
		Status: "released",
	}); err == nil {
		t.Fatal("ListUserGames() error = nil, want validation error")
	}
}

func TestListUserGamesFailsWhenStoreFails(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().ListUserGames(mock.Anything, userID, dto.ListUserGamesInput{Limit: 20}).
		Return(nil, 0, errors.New("db is down"))

	if _, err := uc.ListUserGames(context.Background(), dto.ListUserGamesInput{UserID: userID}); err == nil {
		t.Fatal("ListUserGames() error = nil, want error")
	}
}
