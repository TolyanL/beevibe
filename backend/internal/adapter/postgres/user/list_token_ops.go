package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) ListTokenOps(ctx context.Context, userID uuid.UUID, filter dto.ListTokenOpsInput) ([]dto.ProfileTokenOp, int, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListTokenOps")
	defer span.End()

	const sql = `SELECT COUNT(*) OVER (),
	                    o.game_id,
	                    g.slug,
	                    COALESCE(NULLIF(g.title, ''), g.prompt, ''),
	                    o.delta,
	                    o.reason,
	                    o.created_at
	             FROM token_ops o
	             LEFT JOIN games g ON g.id = o.game_id
	             WHERE o.user_id = $1
	             ORDER BY o.created_at DESC
	             LIMIT $2 OFFSET $3`

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql, userID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	ops := make([]dto.ProfileTokenOp, 0)
	total := 0

	for rows.Next() {
		var (
			gameID    *uuid.UUID
			gameSlug  pgtype.Text
			gameLabel string
			delta     int64
			reason    string
			createdAt time.Time
		)

		if err := rows.Scan(&total, &gameID, &gameSlug, &gameLabel, &delta, &reason, &createdAt); err != nil {
			return nil, 0, fmt.Errorf("rows.Scan: %w", err)
		}

		op := dto.ProfileTokenOp{
			GameID:    gameID,
			GameLabel: gameLabel,
			Delta:     delta,
			Reason:    reason,
			CreatedAt: createdAt,
		}

		if gameSlug.Valid {
			op.GameSlug = &gameSlug.String
		}

		ops = append(ops, op)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows.Err: %w", err)
	}

	return ops, total, nil
}
