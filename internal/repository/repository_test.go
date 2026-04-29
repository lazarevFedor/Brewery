package repository_test

import (
	"Brewery/internal/repository"
	"Brewery/migrator"
	"context"
	"fmt"
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
	// testDB - глобальная переменная для хранения подключения к тестовой базе данных, которая будет использоваться во всех тестах.
	testDB *pgxpool.Pool

	// beerRepo - глобальная переменная для хранения экземпляра BeerPostgres, который будет использоваться в тестах для взаимодействия с таблицей пива.
	beerRepo *repository.BeerPostgres

	// ctgRepo - глобальная переменная для хранения экземпляра CategoryPostgres, который будет использоваться в тестах для взаимодействия с таблицей категорий.
	ctgRepo *repository.CategoryPostgres

	// enumRepo - глобальная переменная для хранения экземпляра EnumPostgres, который будет использоваться в тестах для взаимодействия с таблицей перечислений.
	enumRepo *repository.EnumPostgres
)

//nolint:gocritic, gosec
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
	defer dbContainer.Terminate(ctx)
	if err != nil {
		log.Fatalf("Failed to start container: %s", err)
	}

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

	if err = migrator.Up(testDB); err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}

	beerRepo = repository.NewBeerPostgres(testDB)
	ctgRepo = repository.NewCategoryPostgres(testDB)
	enumRepo = repository.NewEnumPostgres(testDB)

	code := m.Run()

	testDB.Close()
	dbContainer.Terminate(ctx)

	os.Exit(code)
}

func cleanDB(t *testing.T, ctx context.Context, tablename string) {
	_, err := testDB.Exec(ctx,
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tablename),
	)
	if err != nil {
		t.Errorf("failed to clean db: %v", err)
	}
}
