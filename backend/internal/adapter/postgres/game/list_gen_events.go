package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) ListGenEvents(ctx context.Context, gameID uuid.UUID, afterSeq int64) ([]domain.GenEvent, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListGenEvents")
	defer span.End()

	const sql = `SELECT game_id, seq, type, payload, created_at
	             FROM gen_events WHERE game_id = $1 AND seq > $2 ORDER BY seq ASC`

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql, gameID, afterSeq)
	if err != nil {
		return nil, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	events := make([]domain.GenEvent, 0)

	for rows.Next() {
		var (
			gameID    uuid.UUID
			seq       int64
			eventType string
			payload   []byte
			createdAt time.Time
		)

		if err := rows.Scan(&gameID, &seq, &eventType, &payload, &createdAt); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		events = append(events, domain.GenEvent{
			GameID:    gameID,
			Seq:       seq,
			Type:      domain.GenEventType(eventType),
			Payload:   payload,
			CreatedAt: createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return events, nil
}
