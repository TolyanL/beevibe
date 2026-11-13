package game

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
	"github.com/google/uuid"
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
	                    COALESCE(g.description, ''),
	                    COALESCE(g.params_json->>'speed', 'normal'),
	                    g.cover_json,
	                    g.created_at,
	                    g.updated_at,
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
		var (
			game      dto.ProfileGame
			coverJSON []byte
		)

		if err := rows.Scan(&total, &game.ID, &game.Prompt, &game.Genre, &game.Status, &game.Visibility,
			&game.Title, &game.Slug, &game.Description, &game.Speed, &coverJSON, &game.CreatedAt, &game.UpdatedAt,
			&game.TokensCost, &game.ModelCalls, &game.Fixes); err != nil {
			return nil, 0, fmt.Errorf("rows.Scan: %w", err)
		}

		if len(coverJSON) > 0 {
			var cover domain.GameCover

			if err := json.Unmarshal(coverJSON, &cover); err != nil {
				return nil, 0, fmt.Errorf("json.Unmarshal: %w", err)
			}

			if len(cover.Palette) >= 2 {
				game.Cover = &cover
			}
		}

		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows.Err: %w", err)
	}

	return games, total, nil
}
