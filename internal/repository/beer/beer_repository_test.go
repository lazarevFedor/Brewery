// Package repository contains layer that manipulates data in database
package repository_test

import (
	"Brewery/internal/config"
	"Brewery/internal/entities"
	repository "Brewery/internal/repository/beer"
	"Brewery/migrator"
	"Brewery/pkg/postgres"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDBConfig struct {
	postgres postgres.Config `env-prefix:"TEST_POSTGRES_"`
}

func TestBeerRepository_Insert(t *testing.T) {
	ctx := context.Background()

	testCfg := &TestDBConfig{}

	testCfg, err := config.FillConfig(testCfg)
	if err != nil {
		t.Fatalf("failed to fill config")
	}

	dbContainer, err := pgContainer.Run(ctx,
		"postgres:17",
		pgContainer.WithDatabase(testCfg.postgres.DB),
		pgContainer.WithUsername(testCfg.postgres.Username),
		pgContainer.WithPassword(testCfg.postgres.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)

	require.NoError(t, err)

	defer func(dbContainer *pgContainer.PostgresContainer, ctx context.Context) {
		err = dbContainer.Terminate(ctx)
		// TODO: handle or log error
		_ = err
	}(dbContainer, ctx)

	args := []string{
		"sslmode=disable",
		fmt.Sprintf("pool_min_conns=%d", testCfg.postgres.MinConns),
		fmt.Sprintf("pool_max_conns=%d", testCfg.postgres.MaxConns),
	}

	connStr, err := dbContainer.ConnectionString(ctx, args...)
	require.NoError(t, err)

	db, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	defer db.Close()

	err = migrator.Up(db)
	require.NoError(t, err, "Миграции должны применяться без ошибок")

	repo := repository.NewBeerPostgres(db)

	t.Run("Insert Beer", func(t *testing.T) {
		testBeer := entities.Beer{
			Name:        "test",
			Rating:      4.7,
			Description: "test description",
			ABV:         4.7,
			IBU:         100,
			City:        "Москва",
			Country:     "Россия",
			Type:        "Lager",
			Category: entities.ProductCategory{
				Name: "Beer",
			},
			Features: []string{"feat1", "feat2", "feat3"},
		}
		err = repo.InsertBeer(ctx, testBeer)
		require.NoError(t, err)

		var beers []entities.Beer

		beers, err = repo.GetBeers(ctx)
		require.NoError(t, err)
		require.Len(t, beers, 1, "Длина слайса должна быть равно 1")
		require.Equal(t, testBeer, beers[0], "Вставленный и полученный товары должны быть равны")
	})
}
