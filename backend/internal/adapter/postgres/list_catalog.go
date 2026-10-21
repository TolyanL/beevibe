package postgres

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const (
	catalogOrderByCool  = `CASE WHEN cool + bad > 0 THEN cool::float / (cool + bad) ELSE 0 END DESC, created_at DESC`
	catalogOrderByHours = `seconds DESC, created_at DESC`
	catalogOrderByFresh = `created_at DESC`
)

func catalogOrderBy(sort string) string {
	switch sort {
	case "cool":
		return catalogOrderByCool
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
	           COALESCE(ra.cool, 0) AS cool,
	           COALESCE(ra.bad, 0) AS bad,
	           COALESCE(pa.sessions, 0) AS sessions,
	           COALESCE(pa.seconds, 0) AS seconds,
	           COALESCE(va.tokens_cost, 0) AS tokens_cost,
	           COALESCE(va.model_calls, 0) AS model_calls
	    FROM games g
	    LEFT JOIN (
	        SELECT game_id,
	               COUNT(*) FILTER (WHERE value = 'cool') AS cool,
	               COUNT(*) FILTER (WHERE value = 'bad') AS bad
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
	       created_at, cool, bad, sessions, seconds, tokens_cost, model_calls
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
			card      dto.CatalogCard
			cool, bad int
			sessions  int
			seconds   int
		)

		if err := rows.Scan(
			&card.ID, &card.Prompt, &card.Genre, &card.MapSize, &card.Style, &card.CreatedAt,
			&cool, &bad, &sessions, &seconds, &card.TokensCost, &card.ModelCalls,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		card.Rating = dto.RatingAggregate{Cool: cool, Bad: bad, Total: cool + bad}
		if card.Rating.Total > 0 {
			card.Rating.CoolPercent = float64(cool) / float64(card.Rating.Total) * 100
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
