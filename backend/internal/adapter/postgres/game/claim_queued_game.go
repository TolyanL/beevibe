package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) ClaimQueuedGame(ctx context.Context) (domain.Game, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ClaimQueuedGame")
	defer span.End()

	const sql = `WITH claimed AS (
	                 SELECT id FROM games
	                 WHERE status = 'queued'
	                 ORDER BY created_at
	                 FOR UPDATE SKIP LOCKED
	                 LIMIT 1
	             )
	             UPDATE games g
	             SET status = 'analyzing', updated_at = NOW()
	             FROM claimed
	             WHERE g.id = claimed.id
	             RETURNING g.id, g.owner_id, g.prompt, g.params_json, g.spec_json, g.config_json,
	                       g.status, g.fail_reason, g.created_at, g.updated_at`

	row := gameRow{}

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql).Scan(row.Dest()...)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return domain.Game{}, domain.ErrNotFound
		}

		return domain.Game{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}
