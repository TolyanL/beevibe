package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) UpdateGameMeta(ctx context.Context, game domain.Game) error {
	ctx, span := tracer.Start(ctx, "adapter postgres UpdateGameMeta")
	defer span.End()

	const sql = `UPDATE games
	             SET title = $1, description = $2, cover_json = NULLIF($3, '')::jsonb, updated_at = $4
	             WHERE id = $5 AND deleted_at IS NULL`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		game.Title, game.Description, coverJSON(game.Cover), game.UpdatedAt, game.ID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
