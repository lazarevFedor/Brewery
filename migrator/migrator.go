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
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("SetDialect: failed to set SQL dialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(pgPool)
	defer db.Close()

	if err := goose.Up(db, "sql"); err != nil {
		return fmt.Errorf("Up: failed to up migrations: %w", err)
	}

	return nil
}

func Down(gpPool *pgxpool.Pool) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("SetDialect: failed to set SQL dialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(gpPool)
	defer db.Close()

	if err := goose.Down(db, "sql"); err != nil {
		return fmt.Errorf("Down: failed to down migrations: %w", err)
	}

	return nil
}
