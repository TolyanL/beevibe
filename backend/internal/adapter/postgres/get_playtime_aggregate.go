package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetPlaytimeAggregate(ctx context.Context, gameID uuid.UUID) (dto.PlaytimeAggregate, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetPlaytimeAggregate")
	defer span.End()

	const sql = `SELECT COUNT(*), COALESCE(SUM(seconds), 0) FROM play_sessions WHERE game_id = $1`

	var sessions, seconds int

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, gameID).Scan(&sessions, &seconds)
	if err != nil {
		return dto.PlaytimeAggregate{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return dto.PlaytimeAggregate{
		Sessions: sessions,
		Seconds:  seconds,
		Hours:    float64(seconds) / 3600,
	}, nil
}
