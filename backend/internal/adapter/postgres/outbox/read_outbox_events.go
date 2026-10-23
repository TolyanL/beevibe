package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) ReadOutboxEvents(ctx context.Context, limit int) ([]domain.Event, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ReadOutboxEvents")
	defer span.End()

	const sql = `SELECT id, topic, key, payload, created_at
	             FROM outbox WHERE published_at IS NULL ORDER BY created_at ASC LIMIT $1`

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("postgres.ReadOutboxEvents: %w", err)
	}

	defer rows.Close()

	events := make([]domain.Event, 0)

	for rows.Next() {
		var (
			id        uuid.UUID
			topic     string
			key       string
			payload   []byte
			createdAt time.Time
		)

		if err := rows.Scan(&id, &topic, &key, &payload, &createdAt); err != nil {
			return nil, fmt.Errorf("postgres.ReadOutboxEvents: %w", err)
		}

		events = append(events, domain.Event{
			ID:        id,
			Topic:     topic,
			Key:       key,
			Payload:   payload,
			CreatedAt: createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.ReadOutboxEvents: %w", err)
	}

	return events, nil
}
