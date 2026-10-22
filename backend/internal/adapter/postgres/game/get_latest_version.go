package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetLatestVersion(ctx context.Context, gameID uuid.UUID) (domain.Version, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetLatestVersion")
	defer span.End()

	const sql = `SELECT id, game_id, number, html, source, note, tokens_cost, model_calls, prompt_version, created_at
	             FROM versions WHERE game_id = $1 ORDER BY number DESC LIMIT 1`

	row := versionRow{}

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, gameID).Scan(row.Dest()...)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return domain.Version{}, domain.ErrVersionNotFound
		}

		return domain.Version{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}
