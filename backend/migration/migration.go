package migration

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var fs embed.FS

func Up(ctx context.Context, dsn string) error {
	db, err := open(ctx, dsn)
	if err != nil {
		return err
	}

	defer db.Close()

	goose.SetBaseFS(fs)
	goose.SetDialect("postgres")

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose.UpContext: %w", err)
	}

	return nil
}

func Down(ctx context.Context, dsn string) error {
	db, err := open(ctx, dsn)
	if err != nil {
		return err
	}

	defer db.Close()

	goose.SetBaseFS(fs)
	goose.SetDialect("postgres")

	if err := goose.DownContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose.DownContext: %w", err)
	}

	return nil
}

func Status(ctx context.Context, dsn string) error {
	db, err := open(ctx, dsn)
	if err != nil {
		return err
	}

	defer db.Close()

	goose.SetBaseFS(fs)
	goose.SetDialect("postgres")

	if err := goose.StatusContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose.StatusContext: %w", err)
	}

	return nil
}

func open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.PingContext: %w", err)
	}

	return db, nil
}
