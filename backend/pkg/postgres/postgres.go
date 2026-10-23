package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DSN            string        `envconfig:"POSTGRES_DSN"             default:"postgres://bee:12345@localhost:5432/bee?sslmode=disable"`
	MaxConns       int32         `envconfig:"POSTGRES_MAX_CONNS"        default:"10"`
	MinConns       int32         `envconfig:"POSTGRES_MIN_CONNS"        default:"2"`
	ConnectTimeout time.Duration `envconfig:"POSTGRES_CONNECT_TIMEOUT" default:"5s"`
}

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var pool *pgxpool.Pool //nolint:gochecknoglobals

const (
	defaultMaxConns = 10
	defaultMinConns = 2
)

func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	} else {
		poolCfg.MaxConns = defaultMaxConns
	}

	if cfg.MinConns > 0 {
		poolCfg.MinConns = cfg.MinConns
	} else {
		poolCfg.MinConns = defaultMinConns
	}

	if cfg.ConnectTimeout > 0 {
		poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	}

	p, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	return p, nil
}

func Init(p *pgxpool.Pool) {
	pool = p
}

func Pool() *pgxpool.Pool {
	return pool
}
