package game

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgtest"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestListUserGamesScansPages(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()
	gameID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args(userID, "ready", 2, 10)).Return(pgtest.Rows(t,
		[]any{5, gameID, "змейка", "snake", "ready", "private", "", "zmeyka-1a2b", createdAt, 320, 6, 1},
		[]any{5, uuid.New(), "кликер", "clicker", "ready", "public", "", "clicker-1a2b", createdAt, 40, 1, 0},
	), nil)

	games, total, err := postgres.ListUserGames(ctx, userID, dto.ListUserGamesInput{Status: "ready", Sort: "created_at", Limit: 2, Offset: 10})
	if err != nil {
		t.Fatalf("ListUserGames() error = %v, want nil", err)
	}

	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}

	if len(games) != 2 {
		t.Fatalf("games = %d, want 2", len(games))
	}

	if games[0].ID != gameID || games[0].Status != "ready" || games[0].TokensCost != 320 || games[0].Fixes != 1 {
		t.Errorf("game = %+v, want ready змейка with cost 320 and 1 fix", games[0])
	}
}

func TestListUserGamesSortsByTokens(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args(userID, "", 20, 0)).Return(pgtest.Rows(t,
		[]any{1, uuid.New(), "змейка", "snake", "ready", "public", "", "zmeyka-1a2b", time.Now().UTC(), 320, 6, 0},
	), nil)

	if _, _, err := postgres.ListUserGames(ctx, userID, dto.ListUserGamesInput{Sort: "tokens_cost", Limit: 20}); err != nil {
		t.Fatalf("ListUserGames() error = %v, want nil", err)
	}
}

func TestListUserGamesWrapsQueryError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("connection reset"))

	if _, _, err := postgres.ListUserGames(ctx, uuid.New(), dto.ListUserGamesInput{}); err == nil {
		t.Fatal("ListUserGames() error = nil, want error")
	}
}

func TestUserGamesOrderBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sort string
		want string
	}{
		{sort: "tokens_cost", want: userGamesOrderByTokens},
		{sort: "created_at", want: userGamesOrderByCreatedAt},
		{sort: "", want: userGamesOrderByCreatedAt},
		{sort: "unknown", want: userGamesOrderByCreatedAt},
	}

	for _, tt := range tests {
		if got := userGamesOrderBy(tt.sort); got != tt.want {
			t.Errorf("userGamesOrderBy(%q) = %q, want %q", tt.sort, got, tt.want)
		}
	}
}
