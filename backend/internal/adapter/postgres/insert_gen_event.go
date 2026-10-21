package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const genEventInsertAttempts = 5

func (p *Postgres) InsertGenEvent(ctx context.Context, event domain.GenEvent) error {
	ctx, span := tracer.Start(ctx, "adapter postgres InsertGenEvent")
	defer span.End()

	const sql = `INSERT INTO gen_events (game_id, seq, type, payload, created_at)
	             SELECT $1, COALESCE(MAX(seq), 0) + 1, $2, $3::jsonb, $4
	             FROM gen_events WHERE game_id = $1
	             RETURNING seq`

	for attempt := range genEventInsertAttempts {
		var seq int64

		err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql,
			event.GameID, string(event.Type), string(event.Payload), event.CreatedAt).Scan(&seq)
		if err != nil {
			if isUniqueViolation(err) && attempt < genEventInsertAttempts-1 {
				continue
			}

			return fmt.Errorf("txOrPool.QueryRow: %w", err)
		}

		return nil
	}

	return domain.ErrInvalidInput
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
