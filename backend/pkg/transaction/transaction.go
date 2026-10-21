package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TolyanL/beevibe/pkg/postgres"
)

type txKey struct{}

var IsUnitTest bool //nolint:gochecknoglobals // sanctioned global: IsUnitTest

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

func TryExtractTX(ctx context.Context) postgres.DBTX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return postgres.Pool()
}
