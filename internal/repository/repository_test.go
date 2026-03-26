package repository_test

import (
	"Brewery/internal/repository"
	"Brewery/migrator"
	"Brewery/pkg/postgres"
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB   *pgxpool.Pool
	beerRepo *repository.BeerPostgres
	ctgRepo  *repository.CategoryPostgres
)

type TestDBConfig struct {
	Postgres postgres.Config `env-prefix:"TEST_POSTGRES_"`
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbContainer, err := pgContainer.Run(ctx,
		"postgres:17",
		pgContainer.WithDatabase("test_db"),
		pgContainer.WithUsername("test_user"),
		pgContainer.WithPassword("test_pswd"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)

	if err != nil {
		log.Fatalf("Failed to start container: %s", err)
	}
	defer dbContainer.Terminate(ctx)

	args := []string{
		"sslmode=disable",
		"pool_min_conns=1",
		"pool_min_conns=2",
	}

	connStr, err := dbContainer.ConnectionString(ctx, args...)
	if err != nil {
		log.Fatalf("Failed create conn string: %s", err)
	}

	testDB, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s", err)
	}

	if err := migrator.Up(testDB); err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}

	beerRepo = repository.NewBeerPostgres(testDB)
	ctgRepo = repository.NewCategoryPostgres(testDB)

	code := m.Run()

	testDB.Close()
	dbContainer.Terminate(ctx)

	os.Exit(code)
}

func cleanDB(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	_, err := testDB.Exec(context.Background(), "TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Errorf("failed to clean db: %v", err)
	}
}
