package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetVersion(ctx context.Context, gameID uuid.UUID, number int) (domain.Version, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetVersion")
	defer span.End()

	const sql = `SELECT id, game_id, number, html, source, note, tokens_cost, model_calls, prompt_version, created_at
	             FROM versions WHERE game_id = $1 AND number = $2`

	row := versionRow{}

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, gameID, number).Scan(row.Dest()...)
	if err != nil {
		if isNoRows(err) {
			return domain.Version{}, domain.ErrVersionNotFound
		}

		return domain.Version{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}
