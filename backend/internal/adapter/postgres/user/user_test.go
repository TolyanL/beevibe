package user

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgtest"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestCreateUserInsertsRow(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	user := domain.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: "hash",
		TokenBalance: 1000,
		CreatedAt:    time.Now().UTC(),
	}

	db.EXPECT().Exec(mock.Anything, mock.Anything,
		pgtest.Args(user.ID, "user@example.com", "", "hash", int64(1000), mock.Anything)).
		Return(pgtest.Tag(1), nil)

	if err := postgres.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}
}

func TestCreateUserMapsUniqueViolationToEmailTaken(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.Tag(0), pgtest.UniqueViolation("users_email_key"))

	err := postgres.CreateUser(ctx, domain.User{})
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("CreateUser() error = %v, want ErrEmailTaken", err)
	}
}

func TestCreateUserMapsZeroRowsToNotFound(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Tag(0), nil)

	err := postgres.CreateUser(ctx, domain.User{})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("CreateUser() error = %v, want ErrNotFound", err)
	}
}

func TestCreateUserWrapsDriverError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.Tag(0), errors.New("connection reset"))

	if err := postgres.CreateUser(ctx, domain.User{}); err == nil {
		t.Fatal("CreateUser() error = nil, want error")
	}
}

func TestGetUserByIDScansRow(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, pgtest.Args(userID)).
		Return(pgtest.Row(t, userID, "user@example.com", "", "hash", int64(1500), createdAt))

	user, err := postgres.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v, want nil", err)
	}

	if user.ID != userID || user.Email != "user@example.com" || user.TokenBalance != 1500 {
		t.Errorf("user = %+v, want id %s email user@example.com balance 1500", user, userID)
	}

	if !user.CreatedAt.Equal(createdAt) {
		t.Errorf("created at = %v, want %v", user.CreatedAt, createdAt)
	}
}

func TestGetUserByIDMapsNoRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, pgtest.Args(userID)).
		Return(pgtest.RowError(t, pgx.ErrNoRows, 6))

	_, err := postgres.GetUserByID(ctx, userID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetUserByID() error = %v, want ErrNotFound", err)
	}
}

func TestGetSessionUserIDReturnsOwner(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, pgtest.Args("session-token")).
		Return(pgtest.Row(t, userID))

	got, err := postgres.GetSessionUserID(ctx, "session-token")
	if err != nil {
		t.Fatalf("GetSessionUserID() error = %v, want nil", err)
	}

	if got != userID {
		t.Errorf("user id = %s, want %s", got, userID)
	}
}

func TestUpdateUserChecksRowsAffected(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Exec(mock.Anything, mock.Anything, pgtest.Args(
		"user@example.com", "", "hash", int64(1000), mock.Anything,
	)).Return(pgtest.Tag(1), nil)

	if err := postgres.UpdateUser(ctx, domain.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: "hash",
		TokenBalance: 1000,
	}); err != nil {
		t.Fatalf("UpdateUser() error = %v, want nil", err)
	}
}

func TestUpdateUserMapsZeroRowsToNotFound(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Tag(0), nil)

	err := postgres.UpdateUser(ctx, domain.User{})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("UpdateUser() error = %v, want ErrNotFound", err)
	}
}

func TestListTokenOpsScansRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()
	gameID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args(userID, 50, 0)).Return(pgtest.Rows(t,
		[]any{2, uuid.New(), userID, &gameID, int64(-320), "generate", createdAt},
		[]any{2, uuid.New(), userID, nil, int64(500), "topup", createdAt},
	), nil)

	ops, total, err := postgres.ListTokenOps(ctx, userID, dto.ListTokenOpsInput{Limit: 50})
	if err != nil {
		t.Fatalf("ListTokenOps() error = %v, want nil", err)
	}

	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}

	if len(ops) != 2 {
		t.Fatalf("ops = %d, want 2", len(ops))
	}

	if ops[0].GameID == nil || *ops[0].GameID != gameID || ops[0].Reason != domain.ReasonGenerate {
		t.Errorf("first op = %+v, want game %s reason generate", ops[0], gameID)
	}

	if ops[1].GameID != nil || ops[1].Reason != domain.ReasonTopup {
		t.Errorf("second op = %+v, want no game and reason topup", ops[1])
	}
}

func TestListTokenOpsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Rows(t), nil)

	ops, total, err := postgres.ListTokenOps(ctx, userID, dto.ListTokenOpsInput{Limit: 50})
	if err != nil {
		t.Fatalf("ListTokenOps() error = %v, want nil", err)
	}

	if ops == nil || len(ops) != 0 {
		t.Fatalf("ops = %v, want empty slice", ops)
	}

	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

func TestListTokenOpsWrapsQueryError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("connection reset"))

	if _, _, err := postgres.ListTokenOps(ctx, uuid.New(), dto.ListTokenOpsInput{Limit: 50}); err == nil {
		t.Fatal("ListTokenOps() error = nil, want error")
	}
}

func TestCreateSessionInsertsRow(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	userID := uuid.New()
	expiresAt := time.Now().UTC().Add(time.Hour)

	db.EXPECT().Exec(mock.Anything, mock.Anything, pgtest.Args("session-token", userID, expiresAt)).
		Return(pgtest.Tag(1), nil)

	if err := postgres.CreateSession(ctx, "session-token", userID, expiresAt); err != nil {
		t.Fatalf("CreateSession() error = %v, want nil", err)
	}
}

func TestInsertTokenOpStoresOperation(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	op := domain.TokenOp{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		GameID:    &gameID,
		Delta:     -320,
		Reason:    domain.ReasonGenerate,
		CreatedAt: time.Now().UTC(),
	}

	db.EXPECT().Exec(mock.Anything, mock.Anything, pgtest.Args(
		op.ID, op.UserID, op.GameID, op.Delta, string(op.Reason), mock.Anything,
	)).Return(pgtest.Tag(1), nil)

	if err := postgres.InsertTokenOp(ctx, op); err != nil {
		t.Fatalf("InsertTokenOp() error = %v, want nil", err)
	}
}

func TestGetIdempotencyResultReturnsStoredID(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, pgtest.Args("key-1", "create_game")).
		Return(pgtest.Row(t, "game-id"))

	result, found, err := postgres.GetIdempotencyResult(ctx, "key-1", "create_game")
	if err != nil {
		t.Fatalf("GetIdempotencyResult() error = %v, want nil", err)
	}

	if !found || result != "game-id" {
		t.Errorf("result = %q found = %v, want game-id and true", result, found)
	}
}

func TestGetIdempotencyResultReportsMiss(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.RowError(t, pgx.ErrNoRows, 1))

	result, found, err := postgres.GetIdempotencyResult(ctx, "key-1", "create_game")
	if err != nil {
		t.Fatalf("GetIdempotencyResult() error = %v, want nil", err)
	}

	if found || result != "" {
		t.Errorf("result = %q found = %v, want empty and false", result, found)
	}
}
