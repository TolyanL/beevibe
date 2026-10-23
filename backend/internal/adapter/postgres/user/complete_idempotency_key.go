package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) CompleteIdempotencyKey(ctx context.Context, key, operation, resultID string) error {
	ctx, span := tracer.Start(ctx, "adapter postgres CompleteIdempotencyKey")
	defer span.End()

	const sql = `UPDATE idempotency_keys SET result_id = $3 WHERE key = $1 AND operation = $2`

	if _, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, key, operation, resultID); err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	return nil
}
