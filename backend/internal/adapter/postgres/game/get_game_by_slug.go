package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetGameBySlug(ctx context.Context, slug string) (domain.Game, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetGameBySlug")
	defer span.End()

	const sql = `SELECT id, owner_id, prompt, params_json, spec_json, config_json, status, fail_reason,
	                    visibility, deleted_at, title, description, cover_json, slug, created_at, updated_at
	             FROM games WHERE slug = $1 AND deleted_at IS NULL`

	row := gameRow{}

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, slug).Scan(row.Dest()...)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return domain.Game{}, domain.ErrNotFound
		}

		return domain.Game{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}
