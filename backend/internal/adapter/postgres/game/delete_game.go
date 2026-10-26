package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) DeleteGame(ctx context.Context, game domain.Game) error {
	ctx, span := tracer.Start(ctx, "adapter postgres DeleteGame")
	defer span.End()

	const sql = `UPDATE games SET deleted_at = $1, updated_at = $1
	             WHERE id = $2 AND deleted_at IS NULL`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, game.DeletedAt, game.ID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
