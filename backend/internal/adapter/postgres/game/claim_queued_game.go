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
	                 WHERE status = 'queued' AND deleted_at IS NULL
	                 ORDER BY created_at
	                 FOR UPDATE SKIP LOCKED
	                 LIMIT 1
	             )
	             UPDATE games g
	             SET status = 'analyzing', updated_at = NOW()
	             FROM claimed
	             WHERE g.id = claimed.id
	             RETURNING g.id, g.owner_id, g.prompt, g.params_json, g.spec_json, g.config_json,
	                       g.status, g.fail_reason, g.visibility, g.deleted_at,
	                       g.title, g.description, g.cover_json, g.slug, g.created_at, g.updated_at,
	                       g.resume_stage`

	row := gameRow{}

	var resumeStage string

	dest := append(row.Dest(), &resumeStage)

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql).Scan(dest...)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return domain.Game{}, domain.ErrNotFound
		}

		return domain.Game{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	game, err := row.ToDomain()
	if err != nil {
		return domain.Game{}, fmt.Errorf("row.ToDomain: %w", err)
	}

	game.ResumeStage = domain.Status(resumeStage)

	return game, nil
}
