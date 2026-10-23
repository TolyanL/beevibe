package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) ListVersions(ctx context.Context, gameID uuid.UUID) ([]domain.Version, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListVersions")
	defer span.End()

	const sql = `SELECT id, game_id, number, html, source, note, tokens_cost, model_calls, prompt_version, created_at
	             FROM versions WHERE game_id = $1 ORDER BY number ASC`

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql, gameID)
	if err != nil {
		return nil, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	versions := make([]domain.Version, 0)

	for rows.Next() {
		row := versionRow{}

		if err := rows.Scan(row.Dest()...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		version, err := row.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("row.ToDomain: %w", err)
		}

		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return versions, nil
}
