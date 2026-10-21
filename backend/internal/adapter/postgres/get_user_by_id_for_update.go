package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetUserByIDForUpdate(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetUserByIDForUpdate")
	defer span.End()

	const sql = `SELECT id, email, password_hash, token_balance, created_at
	             FROM users WHERE id = $1 FOR UPDATE`

	row := userRow{}

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, userID).Scan(row.Dest()...)
	if err != nil {
		if isNoRows(err) {
			return domain.User{}, domain.ErrNotFound
		}

		return domain.User{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}
