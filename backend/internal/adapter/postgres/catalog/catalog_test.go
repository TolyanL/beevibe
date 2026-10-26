package catalog

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgtest"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestListCatalogBuildsRatingAndPlaytime(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args("snake", "", 20, 0)).Return(pgtest.Rows(t,
		[]any{1, gameID, "змейка", "Змейка", "описание", nil, "zmeyka-1a2b", "bee", uuid.New(),
			"snake", "", "", "medium", "neon", createdAt, 4.5, 4, []int32{0, 0, 1, 2, 1}, 2, 1800, 320, 6},
	), nil)

	cards, total, err := postgres.ListCatalog(ctx, dto.CatalogFilter{Genre: "snake", Limit: 20})
	if err != nil {
		t.Fatalf("ListCatalog() error = %v, want nil", err)
	}

	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}

	if len(cards) != 1 {
		t.Fatalf("cards = %d, want 1", len(cards))
	}

	card := cards[0]

	if card.ID != gameID || card.Genre != "snake" || card.MapSize != "medium" || card.Style != "neon" {
		t.Errorf("card = %+v, want snake medium neon", card)
	}

	if card.Rating.Average != 4.5 || card.Rating.Total != 4 {
		t.Errorf("rating = %+v, want average 4.5 total 4", card.Rating)
	}

	if card.Rating.Counts != [5]int{0, 0, 1, 2, 1} {
		t.Errorf("rating counts = %v, want [0 0 1 2 1]", card.Rating.Counts)
	}

	if card.Playtime.Sessions != 2 || card.Playtime.Seconds != 1800 || card.Playtime.Hours != 0.5 {
		t.Errorf("playtime = %+v, want 2 sessions 1800s 0.5h", card.Playtime)
	}

	if card.TokensCost != 320 || card.ModelCalls != 6 {
		t.Errorf("cost = %d calls = %d, want 320 and 6", card.TokensCost, card.ModelCalls)
	}
}

func TestListCatalogKeepsRatingEmptyWithoutVotes(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Rows(t,
		[]any{1, uuid.New(), "змейка", "", "", nil, "zmeyka-1a2b", "bee", uuid.New(),
			"", "", "", "", "", time.Now().UTC(), 0.0, 0, []int32{0, 0, 0, 0, 0}, 0, 0, 0, 0},
	), nil)

	cards, _, err := postgres.ListCatalog(ctx, dto.CatalogFilter{Limit: 20})
	if err != nil {
		t.Fatalf("ListCatalog() error = %v, want nil", err)
	}

	if cards[0].Rating.Total != 0 || cards[0].Rating.Average != 0 || cards[0].Rating.Counts != [5]int{} {
		t.Errorf("rating = %+v, want empty", cards[0].Rating)
	}
}

func TestListCatalogWrapsQueryError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("connection reset"))

	if _, _, err := postgres.ListCatalog(ctx, dto.CatalogFilter{}); err == nil {
		t.Fatal("ListCatalog() error = nil, want error")
	}
}

func TestCatalogOrderBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sort string
		want string
	}{
		{sort: "rating", want: catalogOrderByRating},
		{sort: "hours", want: catalogOrderByHours},
		{sort: "fresh", want: catalogOrderByFresh},
		{sort: "", want: catalogOrderByFresh},
		{sort: "unknown", want: catalogOrderByFresh},
	}

	for _, tt := range tests {
		if got := catalogOrderBy(tt.sort); got != tt.want {
			t.Errorf("catalogOrderBy(%q) = %q, want %q", tt.sort, got, tt.want)
		}
	}
}

func TestListCatalogFiltersByPromptSubstring(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args("", "змей", 20, 0)).Return(pgtest.Rows(t,
		[]any{2, uuid.New(), "неоновая змейка", "", "", nil, "neon-snake-1a2b", "bee", uuid.New(),
			"snake", "", "", "small", "neon", time.Now().UTC(), 0.0, 0, []int32{0, 0, 0, 0, 0}, 0, 0, 0, 0},
	), nil)

	cards, total, err := postgres.ListCatalog(ctx, dto.CatalogFilter{Q: "змей", Limit: 20})
	if err != nil {
		t.Fatalf("ListCatalog() error = %v, want nil", err)
	}

	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}

	if len(cards) != 1 || cards[0].Prompt != "неоновая змейка" {
		t.Errorf("cards = %+v, want one змейка card", cards)
	}
}
