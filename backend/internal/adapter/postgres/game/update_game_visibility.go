package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) UpdateGameVisibility(ctx context.Context, game domain.Game) error {
	ctx, span := tracer.Start(ctx, "adapter postgres UpdateGameVisibility")
	defer span.End()

	const sql = `UPDATE games SET visibility = $1, updated_at = $2
	             WHERE id = $3 AND deleted_at IS NULL`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, string(game.Visibility), game.UpdatedAt, game.ID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
