package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) InsertPlaySession(ctx context.Context, session domain.PlaySession) error {
	ctx, span := tracer.Start(ctx, "adapter postgres InsertPlaySession")
	defer span.End()

	const sql = `INSERT INTO play_sessions (id, game_id, user_id, seconds, created_at)
	             VALUES ($1, $2, $3, $4, $5)`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		session.ID, session.GameID, session.UserID, session.Seconds, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
