package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) GetRatingAggregate(ctx context.Context, gameID uuid.UUID) (dto.RatingAggregate, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetRatingAggregate")
	defer span.End()

	const sql = `SELECT COUNT(*) FILTER (WHERE value = 'cool'), COUNT(*) FILTER (WHERE value = 'bad')
	             FROM ratings WHERE game_id = $1`

	var cool, bad int

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, gameID).Scan(&cool, &bad)
	if err != nil {
		return dto.RatingAggregate{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	total := cool + bad

	aggregate := dto.RatingAggregate{
		Cool:  cool,
		Bad:   bad,
		Total: total,
	}

	if total > 0 {
		aggregate.CoolPercent = float64(cool) / float64(total) * 100
	}

	return aggregate, nil
}
