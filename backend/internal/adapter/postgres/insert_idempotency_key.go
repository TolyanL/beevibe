package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) InsertIdempotencyKey(ctx context.Context, key, operation string, userID uuid.UUID) (bool, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres InsertIdempotencyKey")
	defer span.End()

	const sql = `INSERT INTO idempotency_keys (key, operation, user_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, key, operation, userID)
	if err != nil {
		return false, fmt.Errorf("txOrPool.Exec: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}
