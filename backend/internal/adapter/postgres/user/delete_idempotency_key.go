package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) DeleteIdempotencyKey(ctx context.Context, key, operation string) error {
	ctx, span := tracer.Start(ctx, "adapter postgres DeleteIdempotencyKey")
	defer span.End()

	const sql = `DELETE FROM idempotency_keys WHERE key = $1 AND operation = $2`

	if _, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, key, operation); err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	return nil
}
