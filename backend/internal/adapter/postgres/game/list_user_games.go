package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const (
	userGamesOrderByCreatedAt = `g.created_at DESC`
	userGamesOrderByTokens    = `tokens_cost DESC, g.created_at DESC`
)

func userGamesOrderBy(sort string) string {
	if sort == "tokens_cost" {
		return userGamesOrderByTokens
	}

	return userGamesOrderByCreatedAt
}

func (p *Postgres) ListUserGames(ctx context.Context, userID uuid.UUID, filter dto.ListUserGamesInput) ([]dto.ProfileGame, int, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListUserGames")
	defer span.End()

	const sql = `SELECT COUNT(*) OVER (),
	                    g.id,
	                    g.prompt,
	                    COALESCE(g.params_json->>'genre', ''),
	                    g.status,
	                    g.visibility,
	                    g.title,
	                    g.slug,
	                    g.created_at,
	                    COALESCE(v.tokens_cost, 0),
	                    COALESCE(v.model_calls, 0),
	                    COALESCE(v.fixes, 0)
	             FROM games g
	             LEFT JOIN (
	                 SELECT game_id,
	                        SUM(tokens_cost) AS tokens_cost,
	                        SUM(model_calls) AS model_calls,
	                        COUNT(*) FILTER (WHERE source = 'repair') AS fixes
	                 FROM versions GROUP BY game_id
	             ) v ON v.game_id = g.id
	             WHERE g.owner_id = $1 AND g.deleted_at IS NULL AND ($2 = '' OR g.status = $2)
	             ORDER BY `

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql+userGamesOrderBy(filter.Sort)+` LIMIT $3 OFFSET $4`,
		userID, filter.Status, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	games := make([]dto.ProfileGame, 0)
	total := 0

	for rows.Next() {
		var game dto.ProfileGame

		if err := rows.Scan(&total, &game.ID, &game.Prompt, &game.Genre, &game.Status, &game.Visibility,
			&game.Title, &game.Slug, &game.CreatedAt,
			&game.TokensCost, &game.ModelCalls, &game.Fixes); err != nil {
			return nil, 0, fmt.Errorf("rows.Scan: %w", err)
		}

		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows.Err: %w", err)
	}

	return games, total, nil
}
