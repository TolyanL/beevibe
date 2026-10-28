package game

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgtest"
	"github.com/TolyanL/beevibe/internal/domain"
)

func TestInsertGameJobWritesRow(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	job, err := domain.NewGameJob(uuid.New(), uuid.New(), domain.JobKindPatch, "сделай фон темнее")
	if err != nil {
		t.Fatalf("NewGameJob() error = %v", err)
	}

	db.EXPECT().Exec(mock.Anything, mock.Anything,
		pgtest.Args(job.ID, job.GameID, job.OwnerID, "patch", "сделай фон темнее", job.CreatedAt),
	).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)

	if err := postgres.InsertGameJob(ctx, job); err != nil {
		t.Fatalf("InsertGameJob() error = %v, want nil", err)
	}
}

func TestInsertGameJobMapsDuplicateToAlreadyExists(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	job, err := domain.NewGameJob(uuid.New(), uuid.New(), domain.JobKindFix, "")
	if err != nil {
		t.Fatalf("NewGameJob() error = %v", err)
	}

	db.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).
		Return(pgconn.CommandTag{}, &pgconn.PgError{Code: "23505"})

	if err := postgres.InsertGameJob(ctx, job); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("InsertGameJob() error = %v, want ErrAlreadyExists", err)
	}
}

func TestClaimGameJobReturnsClaim(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	jobID := uuid.New()
	gameID := uuid.New()
	ownerID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything).Return(pgtest.Row(t,
		jobID, gameID, ownerID, "fix", "игра падает", createdAt,
	))

	job, err := postgres.ClaimGameJob(ctx)
	if err != nil {
		t.Fatalf("ClaimGameJob() error = %v, want nil", err)
	}

	if job.ID != jobID || job.GameID != gameID || job.OwnerID != ownerID {
		t.Errorf("job = %+v, want ids %s/%s/%s", job, jobID, gameID, ownerID)
	}

	if job.Kind != domain.JobKindFix || job.Instruction != "игра падает" {
		t.Errorf("job = %+v, want fix job with instruction", job)
	}

	if !job.CreatedAt.Equal(createdAt) {
		t.Errorf("createdAt = %v, want %v", job.CreatedAt, createdAt)
	}
}

func TestClaimGameJobMapsEmptyQueueToNotFound(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().QueryRow(mock.Anything, mock.Anything).Return(pgtest.RowError(t, pgx.ErrNoRows, 6))

	if _, err := postgres.ClaimGameJob(ctx); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ClaimGameJob() error = %v, want ErrNotFound", err)
	}
}
