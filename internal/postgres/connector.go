// Package postgres contains tools to work with postgres db
package postgres

import(
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"context"
)

type Config struct{
	Host     string `env:"HOST"`
	Port     int    `env:"PORT"`
	DB       string `env:"DB"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	MaxConns int    `env:"MAXCONNS"`
	MinConns int    `env:"MINCONNS"`
}

func NewPool(ctx context.Context, cfg Config)(*pgxpool.Pool, error){
	// urlExample := "postgres://username:password@localhost:5432/database_name?sslmode=disable&pool_min_conns=%d&pool_max_conns=%d"
	connstring := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_min_conns=%d&pool_max_conns=%d",
			cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
		cfg.MinConns,
		cfg.MaxConns,
	)

	pgPool, err := pgxpool.New(ctx, connstring)
	if err != nil{
		return nil, fmt.Errorf("New: failed to create pool: %w", err)
	}

	return pgPool, nil
}