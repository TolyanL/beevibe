package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) UpdateUser(ctx context.Context, user domain.User) error {
	ctx, span := tracer.Start(ctx, "adapter postgres UpdateUser")
	defer span.End()

	const sql = `UPDATE users SET email = $1, display_name = $2, password_hash = $3, token_balance = $4
	             WHERE id = $5`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		string(user.Email), user.DisplayName, string(user.PasswordHash), user.TokenBalance, user.ID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
