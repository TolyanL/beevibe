package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) ListUserGamesWithCost(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.ProfileGame, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListUserGamesWithCost")
	defer span.End()

	const sql = `SELECT g.id,
	                    g.prompt,
	                    COALESCE(g.params_json->>'genre', ''),
	                    g.status,
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
	             WHERE g.owner_id = $1
	             ORDER BY g.created_at DESC
	             LIMIT $2 OFFSET $3`

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	games := make([]dto.ProfileGame, 0)

	for rows.Next() {
		var game dto.ProfileGame

		if err := rows.Scan(&game.ID, &game.Prompt, &game.Genre, &game.Status, &game.CreatedAt,
			&game.TokensCost, &game.ModelCalls, &game.Fixes); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return games, nil
}
