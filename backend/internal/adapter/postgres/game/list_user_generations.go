package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

type userGenerationRow struct {
	Total         int
	GameID        uuid.UUID
	GameSlug      string
	GameTitle     string
	GamePrompt    string
	GameStatus    string
	VersionNumber int
	Source        string
	Note          string
	TokensCost    int
	ModelCalls    int
	CreatedAt     time.Time
}

func (r *userGenerationRow) Dest() []any {
	return []any{
		&r.Total, &r.GameID, &r.GameSlug, &r.GameTitle, &r.GamePrompt, &r.GameStatus,
		&r.VersionNumber, &r.Source, &r.Note, &r.TokensCost, &r.ModelCalls, &r.CreatedAt,
	}
}

func (r *userGenerationRow) ToDTO() dto.ProfileGeneration {
	return dto.ProfileGeneration{
		GameID:        r.GameID,
		GameSlug:      r.GameSlug,
		GameTitle:     r.GameTitle,
		GamePrompt:    r.GamePrompt,
		GameStatus:    r.GameStatus,
		VersionNumber: r.VersionNumber,
		Source:        r.Source,
		Note:          r.Note,
		TokensCost:    r.TokensCost,
		ModelCalls:    r.ModelCalls,
		CreatedAt:     r.CreatedAt,
	}
}

func (p *Postgres) ListUserGenerations(ctx context.Context, userID uuid.UUID, filter dto.ListUserGenerationsInput) ([]dto.ProfileGeneration, int, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListUserGenerations")
	defer span.End()

	const sql = `SELECT COUNT(*) OVER (),
	                    g.id,
	                    g.slug,
	                    g.title,
	                    g.prompt,
	                    g.status,
	                    v.number,
	                    v.source,
	                    v.note,
	                    v.tokens_cost,
	                    v.model_calls,
	                    v.created_at
	             FROM versions v
	             JOIN games g ON g.id = v.game_id
	             WHERE g.owner_id = $1 AND g.deleted_at IS NULL
	             ORDER BY v.created_at DESC
	             LIMIT $2 OFFSET $3`

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql, userID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	generations := make([]dto.ProfileGeneration, 0)
	total := 0

	for rows.Next() {
		row := userGenerationRow{}

		if err := rows.Scan(row.Dest()...); err != nil {
			return nil, 0, fmt.Errorf("rows.Scan: %w", err)
		}

		total = row.Total

		generations = append(generations, row.ToDTO())
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows.Err: %w", err)
	}

	return generations, total, nil
}
