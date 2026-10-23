package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) DeleteSession(ctx context.Context, token string) error {
	ctx, span := tracer.Start(ctx, "adapter postgres DeleteSession")
	defer span.End()

	const sql = `DELETE FROM sessions WHERE token = $1`

	if _, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, token); err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	return nil
}
