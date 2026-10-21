package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) CreateGame(ctx context.Context, game domain.Game) error {
	ctx, span := tracer.Start(ctx, "adapter postgres CreateGame")
	defer span.End()

	params, err := json.Marshal(game.Params)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	const sql = `INSERT INTO games (id, owner_id, prompt, params_json, spec_json, config_json, status, fail_reason, created_at, updated_at)
	             VALUES ($1, $2, $3, $4::jsonb, NULLIF($5, '')::jsonb, NULLIF($6, '')::jsonb, $7, $8, $9, $10)`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		game.ID, game.OwnerID, game.Prompt, string(params), specJSON(game.Spec), configJSON(game.Config),
		game.Status, game.FailReason, game.CreatedAt, game.UpdatedAt)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
