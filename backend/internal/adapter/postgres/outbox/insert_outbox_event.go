package outbox

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) InsertOutboxEvent(ctx context.Context, event domain.Event) error {
	ctx, span := tracer.Start(ctx, "adapter postgres InsertOutboxEvent")
	defer span.End()

	const sql = `INSERT INTO outbox (id, topic, key, payload, created_at)
	             VALUES ($1, $2, $3, $4::jsonb, $5)`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		event.ID, event.Topic, event.Key, string(event.Payload), event.CreatedAt)
	if err != nil {
		return fmt.Errorf("postgres.InsertOutboxEvent: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
