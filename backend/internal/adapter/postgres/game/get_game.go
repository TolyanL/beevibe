package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgerr"
	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

type gameRow struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Prompt     string
	ParamsJSON []byte
	SpecJSON   []byte
	ConfigJSON []byte
	Status     string
	FailReason string
	Visibility  string
	DeletedAt   *time.Time
	Title       string
	Description string
	CoverJSON   []byte
	Slug        string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r *gameRow) Dest() []any {
	return []any{
		&r.ID, &r.OwnerID, &r.Prompt, &r.ParamsJSON, &r.SpecJSON, &r.ConfigJSON,
		&r.Status, &r.FailReason, &r.Visibility, &r.DeletedAt,
		&r.Title, &r.Description, &r.CoverJSON, &r.Slug, &r.CreatedAt, &r.UpdatedAt,
	}
}

func (r *gameRow) ToDomain() (domain.Game, error) {
	game := domain.Game{
		ID:         r.ID,
		OwnerID:    r.OwnerID,
		Prompt:     r.Prompt,
		Status:     domain.Status(r.Status),
		FailReason: r.FailReason,
		Visibility:  domain.Visibility(r.Visibility),
		DeletedAt:   r.DeletedAt,
		Title:       r.Title,
		Description: r.Description,
		Slug:        r.Slug,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}

	if len(r.ParamsJSON) > 0 {
		if err := json.Unmarshal(r.ParamsJSON, &game.Params); err != nil {
			return domain.Game{}, fmt.Errorf("json.Unmarshal: %w", err)
		}
	}

	if len(r.SpecJSON) > 0 {
		var spec domain.GameSpec

		if err := json.Unmarshal(r.SpecJSON, &spec); err != nil {
			return domain.Game{}, fmt.Errorf("json.Unmarshal: %w", err)
		}

		game.Spec = &spec
	}

	if len(r.CoverJSON) > 0 {
		var cover domain.GameCover

		if err := json.Unmarshal(r.CoverJSON, &cover); err != nil {
			return domain.Game{}, fmt.Errorf("json.Unmarshal: %w", err)
		}

		game.Cover = &cover
	}

	if len(r.ConfigJSON) > 0 {
		var cfg domain.GameConfig

		if err := json.Unmarshal(r.ConfigJSON, &cfg); err != nil {
			return domain.Game{}, fmt.Errorf("json.Unmarshal: %w", err)
		}

		game.Config = &cfg
	}

	return game, nil
}

func (p *Postgres) GetGame(ctx context.Context, gameID uuid.UUID) (domain.Game, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetGame")
	defer span.End()

	const sql = `SELECT id, owner_id, prompt, params_json, spec_json, config_json, status, fail_reason,
	                    visibility, deleted_at, title, description, cover_json, slug, created_at, updated_at
	             FROM games WHERE id = $1 AND deleted_at IS NULL`

	row := gameRow{}

	err := transaction.TryExtractTX(ctx).QueryRow(ctx, sql, gameID).Scan(row.Dest()...)
	if err != nil {
		if pgerr.IsNoRows(err) {
			return domain.Game{}, domain.ErrNotFound
		}

		return domain.Game{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}
