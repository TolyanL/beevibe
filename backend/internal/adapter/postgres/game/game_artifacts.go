package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (p *Postgres) SaveGameArtifact(ctx context.Context, gameID uuid.UUID, kind, content string) error {
	ctx, span := tracer.Start(ctx, "adapter postgres SaveGameArtifact")
	defer span.End()

	const sql = `INSERT INTO game_artifacts (game_id, kind, content, updated_at)
	             VALUES ($1, $2, $3, $4)
	             ON CONFLICT (game_id, kind) DO UPDATE
	             SET content = EXCLUDED.content, updated_at = EXCLUDED.updated_at`

	if _, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, gameID, kind, content, time.Now().UTC()); err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	return nil
}

func (p *Postgres) ListGameArtifacts(ctx context.Context, gameID uuid.UUID) (map[string]string, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres ListGameArtifacts")
	defer span.End()

	const sql = `SELECT kind, content FROM game_artifacts WHERE game_id = $1`

	rows, err := transaction.TryExtractTX(ctx).Query(ctx, sql, gameID)
	if err != nil {
		return nil, fmt.Errorf("txOrPool.Query: %w", err)
	}

	defer rows.Close()

	artifacts := make(map[string]string)

	for rows.Next() {
		var kind, content string

		if err := rows.Scan(&kind, &content); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		artifacts[kind] = content
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return artifacts, nil
}
