package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) InsertGameJob(ctx context.Context, job domain.GameJob) error {
	ctx, span := tracer.Start(ctx, "adapter postgres InsertGameJob")
	defer span.End()

	const sql = `INSERT INTO game_jobs (id, game_id, owner_id, kind, instruction, created_at)
	             VALUES ($1, $2, $3, $4, $5, $6)`

	if _, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		job.ID, job.GameID, job.OwnerID, string(job.Kind), job.Instruction, job.CreatedAt); err != nil {
		if pgerr.IsUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}

		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	return nil
}
