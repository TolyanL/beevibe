package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) SumTopupsSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres SumTopupsSince")
	defer span.End()

	const sql = `SELECT COALESCE(SUM(delta), 0) FROM token_ops
	             WHERE user_id = $1 AND reason = $2 AND created_at >= $3`

	var total int64

	if err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, userID, string(domain.ReasonTopup), since).
		Scan(&total); err != nil {
		return 0, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return total, nil
}
