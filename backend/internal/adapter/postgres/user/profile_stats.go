package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetProfileStats(ctx context.Context, userID uuid.UUID) (dto.ProfileStats, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetProfileStats")
	defer span.End()

	const sql = `SELECT (SELECT COUNT(*) FROM games g WHERE g.owner_id = $1 AND g.deleted_at IS NULL),
	                    (SELECT COUNT(*) FROM games g WHERE g.owner_id = $1 AND g.deleted_at IS NULL AND g.status = 'ready'),
	                    (SELECT COALESCE(-SUM(o.delta) FILTER (WHERE o.reason IN ('generate', 'patch', 'meta', 'refund')), 0)
	                     FROM token_ops o WHERE o.user_id = $1),
	                    (SELECT COUNT(*) FROM versions v
	                     JOIN games g ON g.id = v.game_id
	                     WHERE g.owner_id = $1 AND g.deleted_at IS NULL AND v.source = 'repair')`

	var stats dto.ProfileStats

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, userID).
		Scan(&stats.GamesTotal, &stats.GamesReady, &stats.TokensSpent, &stats.Fixes)
	if err != nil {
		return dto.ProfileStats{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return stats, nil
}
