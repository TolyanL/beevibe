package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"

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

func (p *Postgres) ListCatalog(ctx context.Context, filter dto.CatalogFilter) ([]dto.CatalogCard, int, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListCatalog")
	defer span.End()

	const query = `WITH catalog AS (
	    SELECT g.id,
	           g.prompt,
	           g.title,
	           g.description,
	           g.cover_json,
	           g.slug,
	           COALESCE(u.display_name, '') AS author_name,
	           g.owner_id,
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
	    LEFT JOIN users u ON u.id = g.owner_id
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
	    WHERE g.status = 'ready' AND g.visibility = 'public' AND g.deleted_at IS NULL
	      AND ($1 = '' OR COALESCE(g.params_json->>'genre', '') = $1)
	          AND ($2 = '' OR g.prompt ILIKE '%' || $2 || '%')
	)
	SELECT COUNT(*) OVER (),
	       id, prompt, title, description, cover_json, slug, author_name, owner_id,
	       COALESCE(params_json->>'genre', ''),
	       COALESCE(params_json->>'custom_genre', ''),
	       COALESCE(params_json->>'custom_style', ''),
	       COALESCE(params_json->>'map_size', ''),
	       COALESCE(params_json->>'style', ''),
	       created_at, rating_average, rating_total, rating_counts, sessions, seconds, tokens_cost, model_calls
	FROM catalog
	ORDER BY `

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, query+catalogOrderBy(filter.Sort)+` LIMIT $3 OFFSET $4`,
		filter.Genre, filter.Q, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	cards := make([]dto.CatalogCard, 0)
	total := 0

	for rows.Next() {
		var (
			card          dto.CatalogCard
			coverRaw      []byte
			ownerID       uuid.UUID
			ratingAverage float64
			ratingTotal   int
			ratingCounts  []int32
			sessions      int
			seconds       int
		)

		if err := rows.Scan(
			&total, &card.ID, &card.Prompt, &card.Title, &card.Description, &coverRaw,
			&card.Slug, &card.Author, &ownerID,
			&card.Genre, &card.CustomGenre, &card.CustomStyle, &card.MapSize, &card.Style, &card.CreatedAt,
			&ratingAverage, &ratingTotal, &ratingCounts, &sessions, &seconds, &card.TokensCost, &card.ModelCalls,
		); err != nil {
			return nil, 0, fmt.Errorf("rows.Scan: %w", err)
		}

		if len(coverRaw) > 0 {
			var cover domain.GameCover

			if err := json.Unmarshal(coverRaw, &cover); err == nil {
				card.Cover = &cover
			}
		}

		card.Author = authorName(card.Author, ownerID)
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
		return nil, 0, fmt.Errorf("rows.Err: %w", err)
	}

	return cards, total, nil
}

func authorName(name string, ownerID uuid.UUID) string {
	if name != "" {
		return name
	}

	return "bee-" + ownerID.String()[:8]
}
