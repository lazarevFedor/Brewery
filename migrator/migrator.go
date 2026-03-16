// Package migrator UPs and DOWNs db migrations
package migrator

import (
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

func Up(pgPool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pgPool)
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("SetDialect: failed to set SQL dialect: %W", err)
	}

	if err := goose.Up(db, "sql"); err != nil {
		return fmt.Errorf("Up: failed to up migrations: %W", err)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("Close: failed to close db connection: %w", err)
	}

	return nil
}