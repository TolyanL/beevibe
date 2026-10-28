package game

import (
	"context"
	"fmt"
	"time"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) ClaimGameJob(ctx context.Context) (domain.GameJob, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ClaimGameJob")
	defer span.End()

	const sql = `DELETE FROM game_jobs
	             WHERE id = (
	                 SELECT id FROM game_jobs
	                 ORDER BY created_at
	                 FOR UPDATE SKIP LOCKED
	                 LIMIT 1
	             )
	             RETURNING id, game_id, owner_id, kind, instruction, created_at`

	var (
		job         domain.GameJob
		kind        string
		instruction string
		createdAt   time.Time
	)

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql).
		Scan(&job.ID, &job.GameID, &job.OwnerID, &kind, &instruction, &createdAt)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return domain.GameJob{}, domain.ErrNotFound
		}

		return domain.GameJob{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	job.Kind = domain.JobKind(kind)
	job.Instruction = instruction
	job.CreatedAt = createdAt

	return job, nil
}
