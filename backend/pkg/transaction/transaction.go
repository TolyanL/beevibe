package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TolyanL/beevibe/pkg/postgres"
)

type txKey struct{}

type dbtxKey struct{}

var IsUnitTest bool //nolint:gochecknoglobals

func Init(p *pgxpool.Pool) {
	postgres.Init(p)
}

func Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	if IsUnitTest {
		return fn(ctx)
	}

	tx, err := postgres.Pool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("tx.Rollback: %w", rollbackErr)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

// NOTE: the seam exists for the adapter tests: they put a mocked DBTX into the
// context instead of spawning a real pool.
func WithDBTX(ctx context.Context, db postgres.DBTX) context.Context {
	return context.WithValue(ctx, dbtxKey{}, db)
}

func TryExtractTX(ctx context.Context) postgres.DBTX {
	if db, ok := ctx.Value(dbtxKey{}).(postgres.DBTX); ok {
		return db
	}

	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return postgres.Pool()
}
