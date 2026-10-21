package postgres

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) UpsertRating(ctx context.Context, rating domain.Rating) error {
	ctx, span := tracer.Start(ctx, "adapter postgres UpsertRating")
	defer span.End()

	const sql = `INSERT INTO ratings (game_id, user_id, value, updated_at)
	             VALUES ($1, $2, $3, $4)
	             ON CONFLICT (game_id, user_id) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		rating.GameID, rating.UserID, string(rating.Value), rating.UpdatedAt)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
