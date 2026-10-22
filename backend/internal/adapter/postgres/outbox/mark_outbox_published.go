package outbox

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) MarkOutboxPublished(ctx context.Context, ids []uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "adapter postgres MarkOutboxPublished")
	defer span.End()

	const sql = `UPDATE outbox SET published_at = NOW(), attempts = attempts + 1 WHERE id = ANY($1)`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, ids)
	if err != nil {
		return fmt.Errorf("postgres.MarkOutboxPublished: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
