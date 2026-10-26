package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) CreateUser(ctx context.Context, user domain.User) error {
	ctx, span := tracer.Start(ctx, "adapter postgres CreateUser")
	defer span.End()

	const sql = `INSERT INTO users (id, email, display_name, password_hash, token_balance, created_at)
	             VALUES ($1, $2, $3, $4, $5, $6)`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		user.ID, string(user.Email), user.DisplayName, string(user.PasswordHash), user.TokenBalance, user.CreatedAt)
	if err != nil {
		if pgerr.IsUniqueViolation(err) {
			return domain.ErrEmailTaken
		}

		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
