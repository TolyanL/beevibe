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

func TestListUserGenerationsScansRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()
	gameID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args(userID, 2, 4)).Return(pgtest.Rows(t,
		[]any{7, gameID, "zmeyka-1a2b", "Змейка", "змейка", "ready", 3, "patch", "плавнее", 40, 1, createdAt},
		[]any{7, uuid.New(), "clicker-1a2b", "", "кликер", "ready", 1, "generate", "", 320, 2, createdAt},
	), nil)

	items, total, err := postgres.ListUserGenerations(ctx, userID, dto.ListUserGenerationsInput{Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("ListUserGenerations() error = %v, want nil", err)
	}

	if total != 7 {
		t.Errorf("total = %d, want 7", total)
	}

	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}

	first := items[0]
	if first.GameID != gameID || first.GameSlug != "zmeyka-1a2b" || first.GameTitle != "Змейка" ||
		first.GameStatus != "ready" || first.VersionNumber != 3 || first.Source != "patch" ||
		first.Note != "плавнее" || first.TokensCost != 40 || first.ModelCalls != 1 {
		t.Errorf("first item = %+v, want version 3 patch of %s", first, gameID)
	}

	if items[1].GamePrompt != "кликер" || items[1].Source != "generate" {
		t.Errorf("second item = %+v, want generate of кликер", items[1])
	}
}

func TestListUserGenerationsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Rows(t), nil)

	items, total, err := postgres.ListUserGenerations(ctx, uuid.New(), dto.ListUserGenerationsInput{Limit: 20})
	if err != nil {
		t.Fatalf("ListUserGenerations() error = %v, want nil", err)
	}

	if items == nil || len(items) != 0 {
		t.Fatalf("items = %v, want empty slice", items)
	}

	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

func TestListUserGenerationsWrapsQueryError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("connection reset"))

	if _, _, err := postgres.ListUserGenerations(ctx, uuid.New(), dto.ListUserGenerationsInput{Limit: 20}); err == nil {
		t.Fatal("ListUserGenerations() error = nil, want error")
	}
}

func TestListUserGenerationsWrapsRowsError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.RowsError(t, errors.New("connection reset")), nil)

	if _, _, err := postgres.ListUserGenerations(ctx, uuid.New(), dto.ListUserGenerationsInput{Limit: 20}); err == nil {
		t.Fatal("ListUserGenerations() error = nil, want error")
	}
}
