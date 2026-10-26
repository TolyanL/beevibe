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

	const sql = `SELECT COALESCE(AVG(value), 0)::float8,
	                    COUNT(*),
	                    COUNT(*) FILTER (WHERE value = 1),
	                    COUNT(*) FILTER (WHERE value = 2),
	                    COUNT(*) FILTER (WHERE value = 3),
	                    COUNT(*) FILTER (WHERE value = 4),
	                    COUNT(*) FILTER (WHERE value = 5)
	             FROM ratings WHERE game_id = $1`

	var (
		aggregate          dto.RatingAggregate
		v1, v2, v3, v4, v5 int
	)

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, gameID).
		Scan(&aggregate.Average, &aggregate.Total, &v1, &v2, &v3, &v4, &v5)
	if err != nil {
		return dto.RatingAggregate{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	aggregate.Counts = [5]int{v1, v2, v3, v4, v5}

	return aggregate, nil
}
