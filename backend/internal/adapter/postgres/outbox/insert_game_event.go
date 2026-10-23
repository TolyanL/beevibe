package outbox

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) InsertGameEvent(ctx context.Context, record domain.GameEventRecord) error {
	ctx, span := tracer.Start(ctx, "adapter postgres InsertGameEvent")
	defer span.End()

	const sql = `INSERT INTO game_event_log (id, event_id, topic, key, event_type, payload, received_at)
	             VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
	             ON CONFLICT (event_id) DO NOTHING`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		record.ID, record.EventID, record.Topic, record.Key, string(record.EventType),
		string(record.Payload), record.ReceivedAt)
	if err != nil {
		if pgerr.IsUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}

		return fmt.Errorf("postgres.InsertGameEvent: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrAlreadyExists
	}

	return nil
}
