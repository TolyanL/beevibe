package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) InsertTokenOp(ctx context.Context, op domain.TokenOp) error {
	ctx, span := tracer.Start(ctx, "adapter postgres InsertTokenOp")
	defer span.End()

	const sql = `INSERT INTO token_ops (id, user_id, game_id, delta, reason, created_at)
	             VALUES ($1, $2, $3, $4, $5, $6)`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		op.ID, op.UserID, op.GameID, op.Delta, string(op.Reason), op.CreatedAt)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
