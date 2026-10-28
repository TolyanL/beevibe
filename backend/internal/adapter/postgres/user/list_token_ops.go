package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) ListTokenOps(ctx context.Context, userID uuid.UUID, filter dto.ListTokenOpsInput) ([]domain.TokenOp, int, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListTokenOps")
	defer span.End()

	const sql = `SELECT COUNT(*) OVER (), id, user_id, game_id, delta, reason, created_at
	             FROM token_ops WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql, userID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	ops := make([]domain.TokenOp, 0)
	total := 0

	for rows.Next() {
		var (
			id        uuid.UUID
			ownerID   uuid.UUID
			gameID    *uuid.UUID
			delta     int64
			reason    string
			createdAt time.Time
		)

		if err := rows.Scan(&total, &id, &ownerID, &gameID, &delta, &reason, &createdAt); err != nil {
			return nil, 0, fmt.Errorf("rows.Scan: %w", err)
		}

		ops = append(ops, domain.TokenOp{
			ID:        id,
			UserID:    ownerID,
			GameID:    gameID,
			Delta:     delta,
			Reason:    domain.TokenReason(reason),
			CreatedAt: createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows.Err: %w", err)
	}

	return ops, total, nil
}
