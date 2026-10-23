package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetSessionUserID(ctx context.Context, token string) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetSessionUserID")
	defer span.End()

	const sql = `SELECT user_id FROM sessions WHERE token = $1 AND expires_at > NOW()`

	var userID uuid.UUID

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, token).Scan(&userID)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return uuid.Nil, domain.ErrNotFound
		}

		return uuid.Nil, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return userID, nil
}
