package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) UpdateGameStatus(ctx context.Context, gameID uuid.UUID, status domain.Status, failReason string) error {
	ctx, span := tracer.Start(ctx, "adapter postgres UpdateGameStatus")
	defer span.End()

	const sql = `UPDATE games SET status = $1, fail_reason = $2, updated_at = NOW() WHERE id = $3`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, string(status), failReason, gameID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
