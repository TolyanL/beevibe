package postgres

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetIdempotencyResult(ctx context.Context, key, operation string) (string, bool, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetIdempotencyResult")
	defer span.End()

	const sql = `SELECT result_id FROM idempotency_keys WHERE key = $1 AND operation = $2`

	var resultID string

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, key, operation).Scan(&resultID)
	if err != nil {
		if isNoRows(err) {
			return "", false, nil
		}

		return "", false, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return resultID, true, nil
}
