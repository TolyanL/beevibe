package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetRating(ctx context.Context, gameID, userID uuid.UUID) (domain.Rating, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetRating")
	defer span.End()

	const sql = `SELECT game_id, user_id, value, updated_at FROM ratings WHERE game_id = $1 AND user_id = $2`

	var (
		storedGameID uuid.UUID
		storedUserID uuid.UUID
		value        string
		updatedAt    time.Time
	)

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, gameID, userID).Scan(&storedGameID, &storedUserID, &value, &updatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Rating{}, domain.ErrNotFound
		}

		return domain.Rating{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return domain.Rating{
		GameID:    storedGameID,
		UserID:    storedUserID,
		Value:     domain.RatingValue(value),
		UpdatedAt: updatedAt,
	}, nil
}
