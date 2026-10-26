package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) UpdateDisplayName(ctx context.Context, user domain.User) error {
	ctx, span := tracer.Start(ctx, "adapter postgres UpdateDisplayName")
	defer span.End()

	const sql = `UPDATE users SET display_name = $1 WHERE id = $2`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, user.DisplayName, user.ID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
