package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) CreateSession(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	ctx, span := tracer.Start(ctx, "adapter postgres CreateSession")
	defer span.End()

	const sql = `INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
