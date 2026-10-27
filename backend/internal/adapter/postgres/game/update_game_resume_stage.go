package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) SaveGameResumeStage(ctx context.Context, gameID uuid.UUID, stage domain.Status) error {
	ctx, span := tracer.Start(ctx, "adapter postgres SaveGameResumeStage")
	defer span.End()

	const sql = `UPDATE games SET resume_stage = $1 WHERE id = $2 AND deleted_at IS NULL`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, string(stage), gameID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
