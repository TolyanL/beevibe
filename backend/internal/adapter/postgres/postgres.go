package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TolyanL/beevibe/pkg/tracer"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) Ping(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "adapter postgres Ping")
	defer span.End()

	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("pool.Ping: %w", err)
	}

	return nil
}
