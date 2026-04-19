package main

import (
	"Brewery/migrator"
	"Brewery/pkg/postgres"
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	postgresCfg := postgres.Config{
		Host:     "localhost",
		Port:     5432,
		Username: "user",
		Password: "1234",
		DB:       "brewery_db",
		MinConns: 1,
		MaxConns: 10,
	}
	pool, err := postgres.NewPool(ctx, postgresCfg)
	if err != nil {
		panic(fmt.Errorf("failed to create postgres pool: %w", err))
	}

	if err = migrator.Down(pool); err != nil {
		panic(fmt.Errorf("failed to run migrations: %w", err))
	}

}
