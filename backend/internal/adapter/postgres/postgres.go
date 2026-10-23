package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/catalog"
	"github.com/TolyanL/beevibe/internal/adapter/postgres/game"
	"github.com/TolyanL/beevibe/internal/adapter/postgres/outbox"
	"github.com/TolyanL/beevibe/internal/adapter/postgres/user"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

type (
	UserPostgres    = user.Postgres
	GamePostgres    = game.Postgres
	CatalogPostgres = catalog.Postgres
	OutboxPostgres  = outbox.Postgres
)

type Postgres struct {
	pool *pgxpool.Pool

	*UserPostgres
	*GamePostgres
	*CatalogPostgres
	*OutboxPostgres
}

func New(pool *pgxpool.Pool) *Postgres {
	return &Postgres{
		pool:            pool,
		UserPostgres:    user.New(pool),
		GamePostgres:    game.New(pool),
		CatalogPostgres: catalog.New(pool),
		OutboxPostgres:  outbox.New(pool),
	}
}

func (p *Postgres) Ping(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "adapter postgres Ping")
	defer span.End()

	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("pool.Ping: %w", err)
	}

	return nil
}
