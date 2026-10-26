package catalog

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const (
	catalogOrderByRating = `rating_average DESC, rating_total DESC, created_at DESC`
	catalogOrderByHours  = `seconds DESC, created_at DESC`
	catalogOrderByFresh  = `created_at DESC`
)

func catalogOrderBy(sort string) string {
	switch sort {
	case "rating":
		return catalogOrderByRating
	case "hours":
		return catalogOrderByHours
	default:
		return catalogOrderByFresh
	}
}

func (p *Postgres) ListCatalog(ctx context.Context, filter dto.CatalogFilter) ([]dto.CatalogCard, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListCatalog")
	defer span.End()

	const query = `WITH catalog AS (
	    SELECT g.id,
	           g.prompt,
	           g.params_json,
	           g.created_at,
	           COALESCE(ra.rating_average, 0) AS rating_average,
	           COALESCE(ra.rating_total, 0) AS rating_total,
	           COALESCE(ra.rating_counts, ARRAY[0, 0, 0, 0, 0]::int[]) AS rating_counts,
	           COALESCE(pa.sessions, 0) AS sessions,
	           COALESCE(pa.seconds, 0) AS seconds,
	           COALESCE(va.tokens_cost, 0) AS tokens_cost,
	           COALESCE(va.model_calls, 0) AS model_calls
	    FROM games g
	    LEFT JOIN (
	        SELECT game_id,
	               AVG(value)::float8 AS rating_average,
	               COUNT(*) AS rating_total,
	               ARRAY[
	                   COUNT(*) FILTER (WHERE value = 1),
	                   COUNT(*) FILTER (WHERE value = 2),
	                   COUNT(*) FILTER (WHERE value = 3),
	                   COUNT(*) FILTER (WHERE value = 4),
	                   COUNT(*) FILTER (WHERE value = 5)
	               ]::int[] AS rating_counts
	        FROM ratings GROUP BY game_id
	    ) ra ON ra.game_id = g.id
	    LEFT JOIN (
	        SELECT game_id, COUNT(*) AS sessions, COALESCE(SUM(seconds), 0) AS seconds
	        FROM play_sessions GROUP BY game_id
	    ) pa ON pa.game_id = g.id
	    LEFT JOIN (
	        SELECT game_id, SUM(tokens_cost) AS tokens_cost, SUM(model_calls) AS model_calls
	        FROM versions GROUP BY game_id
	    ) va ON va.game_id = g.id
	    WHERE g.status = 'ready' AND ($1 = '' OR COALESCE(g.params_json->>'genre', '') = $1)
	)
	SELECT id, prompt,
	       COALESCE(params_json->>'genre', ''),
	       COALESCE(params_json->>'map_size', ''),
	       COALESCE(params_json->>'style', ''),
	       created_at, rating_average, rating_total, rating_counts, sessions, seconds, tokens_cost, model_calls
	FROM catalog
	ORDER BY `

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, query+catalogOrderBy(filter.Sort)+` LIMIT $2 OFFSET $3`,
		filter.Genre, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	cards := make([]dto.CatalogCard, 0)

	for rows.Next() {
		var (
			card          dto.CatalogCard
			ratingAverage float64
			ratingTotal   int
			ratingCounts  []int32
			sessions      int
			seconds       int
		)

		if err := rows.Scan(
			&card.ID, &card.Prompt, &card.Genre, &card.MapSize, &card.Style, &card.CreatedAt,
			&ratingAverage, &ratingTotal, &ratingCounts, &sessions, &seconds, &card.TokensCost, &card.ModelCalls,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		card.Rating = dto.RatingAggregate{Average: ratingAverage, Total: ratingTotal}
		for i := 0; i < len(ratingCounts) && i < len(card.Rating.Counts); i++ {
			card.Rating.Counts[i] = int(ratingCounts[i])
		}

		card.Playtime = dto.PlaytimeAggregate{
			Sessions: sessions,
			Seconds:  seconds,
			Hours:    float64(seconds) / 3600,
		}

		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return cards, nil
}
