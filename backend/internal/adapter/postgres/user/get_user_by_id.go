package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

type userRow struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	PasswordHash string
	TokenBalance int64
	CreatedAt    time.Time
}

func (r *userRow) Dest() []any {
	return []any{&r.ID, &r.Email, &r.DisplayName, &r.PasswordHash, &r.TokenBalance, &r.CreatedAt}
}

func (r *userRow) ToDomain() (domain.User, error) {
	return domain.User{
		ID:           r.ID,
		Email:        domain.Email(r.Email),
		DisplayName:  r.DisplayName,
		PasswordHash: domain.PasswordHash(r.PasswordHash),
		TokenBalance: r.TokenBalance,
		CreatedAt:    r.CreatedAt,
	}, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetUserByID")
	defer span.End()

	const sql = `SELECT id, email, display_name, password_hash, token_balance, created_at
	             FROM users WHERE id = $1`

	row := userRow{}

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, userID).Scan(row.Dest()...)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return domain.User{}, domain.ErrNotFound
		}

		return domain.User{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}
