// Package repository_test содержит тесты для слоя repository
package repository_test

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"Brewery/migrator"
	"context"
	"errors"
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

// TestMain запускает тестовую среду с помощью testcontainers, выполняет миграции и очищает ресурсы после тестов.
//
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

	if err = seedTestData(ctx); err != nil {
		log.Fatalf("Failed to seed test data: %s", err)
	}

	code := m.Run()

	testDB.Close()
	if err = dbContainer.Terminate(ctx); err != nil {
		log.Printf("Failed to terminate container: %s", err)
	}

	os.Exit(code)
}

// cleanDB выполняет очистку указанной таблицы в базе данных, удаляя все записи и сбрасывая идентификаторы.
func cleanDB(t *testing.T, ctx context.Context, tablename string) {
	_, err := testDB.Exec(ctx, "TRUNCATE TABLE "+tablename+" RESTART IDENTITY CASCADE")
	if err != nil {
		t.Errorf("failed to clean db: %v", err)
	}
}

// seedTestData выполняет начальную загрузку тестовых данных в базу данных, обеспечивая наличие необходимых записей для тестов.
func seedTestData(ctx context.Context) error {
	countryID, err := beerRepo.GetCountryID(ctx, nil, "test_country")
	if err != nil {
		return fmt.Errorf("seed country: %w", err)
	}

	if countryID == 0 {
		return errors.New("seed country: zero id")
	}

	cityID, err := beerRepo.GetCityID(ctx, nil, "test_city", countryID)
	if err != nil {
		return fmt.Errorf("seed city: %w", err)
	}

	if cityID == 0 {
		return errors.New("seed city: zero id")
	}
	var categoryID uint
	categoryID, err = ctgRepo.GetCategoryID(ctx, nil, "test_category")
	if err != nil {
		if categoryID == 0 {
			categoryID, err = ctgRepo.InsertCategory(ctx, nil, entities.ProductCategory{Name: "test_category"})
			if err != nil {
				return fmt.Errorf("seed category insert: %w", err)
			}
		} else {
			return fmt.Errorf("seed category get id: %w", err)
		}
	}

	if categoryID == 0 {
		return errors.New("seed category: zero id")
	}

	return nil
}
