package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) UpdateGame(ctx context.Context, game domain.Game) error {
	ctx, span := tracer.Start(ctx, "adapter postgres UpdateGame")
	defer span.End()

	params, err := json.Marshal(game.Params)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	const sql = `UPDATE games
	             SET prompt = $1, params_json = $2::jsonb, spec_json = NULLIF($3, '')::jsonb,
	                 config_json = NULLIF($4, '')::jsonb, status = $5, fail_reason = $6, updated_at = $7
	             WHERE id = $8`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		game.Prompt, string(params), specJSON(game.Spec), configJSON(game.Config),
		game.Status, game.FailReason, game.UpdatedAt, game.ID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
