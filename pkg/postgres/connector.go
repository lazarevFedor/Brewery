// Package postgres contains tools to work with postgres db
package postgres

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config описывает переменные и данные, необходимые для работы базой
//
//nolint:gosec
type Config struct {
	Host     string `env:"HOST"`
	Port     int    `env:"PORT"`
	DB       string `env:"DB"`
	Username string `env:"USER"`
	Password string `env:"PASSWORD"`
	MaxConns int    `env:"MAX_CONNS"`
	MinConns int    `env:"MIN_CONNS"`
}

// NewPool создает пул подлючений в бд
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	// urlExample := "postgres://username:password@localhost:5432/database_name?sslmode=disable&pool_min_conns=%d&pool_max_conns=%d"
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	connstring := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&pool_min_conns=%d&pool_max_conns=%d",
		cfg.Username,
		cfg.Password,
		addr,
		cfg.DB,
		cfg.MinConns,
		cfg.MaxConns,
	)

	pgPool, err := pgxpool.New(ctx, connstring)
	if err != nil {
		return nil, fmt.Errorf("new: failed to create pool: %w", err)
	}

	return pgPool, nil
}
