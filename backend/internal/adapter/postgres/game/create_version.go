package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

type versionRow struct {
	ID            uuid.UUID
	GameID        uuid.UUID
	Number        int
	HTML          string
	Source        string
	Note          string
	TokensCost    int
	ModelCalls    int
	PromptVersion string
	CreatedAt     time.Time
}

func (r *versionRow) Dest() []any {
	return []any{
		&r.ID, &r.GameID, &r.Number, &r.HTML, &r.Source, &r.Note,
		&r.TokensCost, &r.ModelCalls, &r.PromptVersion, &r.CreatedAt,
	}
}

func (r *versionRow) ToDomain() (domain.Version, error) {
	return domain.Version{
		ID:            r.ID,
		GameID:        r.GameID,
		Number:        r.Number,
		HTML:          r.HTML,
		Source:        domain.VersionSource(r.Source),
		Note:          r.Note,
		TokensCost:    r.TokensCost,
		ModelCalls:    r.ModelCalls,
		PromptVersion: r.PromptVersion,
		CreatedAt:     r.CreatedAt,
	}, nil
}

func (p *Postgres) CreateVersion(ctx context.Context, version domain.Version) error {
	ctx, span := tracer.Start(ctx, "adapter postgres CreateVersion")
	defer span.End()

	const sql = `INSERT INTO versions (id, game_id, number, html, source, note, tokens_cost, model_calls, prompt_version, created_at)
	             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql,
		version.ID, version.GameID, version.Number, version.HTML, string(version.Source), version.Note,
		version.TokensCost, version.ModelCalls, version.PromptVersion, version.CreatedAt)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
