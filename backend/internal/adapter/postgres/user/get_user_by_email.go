package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetUserByEmail")
	defer span.End()

	const sql = `SELECT id, email, password_hash, token_balance, created_at
	             FROM users WHERE email = $1`

	row := userRow{}

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, email).Scan(row.Dest()...)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return domain.User{}, domain.ErrNotFound
		}

		return domain.User{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}
