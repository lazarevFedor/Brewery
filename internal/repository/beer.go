// Package repository contains layer that manipulates data in database
package repository

import (
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"

	// sq "github.com/Masterminds/squirrel"
	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/get_or_create_country.sql
	getOrCreateCountryQuery string

	//go:embed sql/get_or_create_city.sql
	getOrCreateCityQuery string

	//go:embed sql/get_or_create_type.sql
	getOrCreateTypeQuery string

	//go:embed sql/get_product_category_by_name.sql
	getProductCategoryByNameQuery string

	//go:embed sql/insert_beer.sql
	insertBeerQuery string

	//go:embed sql/insert_feature.sql
	insertFeatureQuery string

	//go:embed sql/insert_beer_feature.sql
	insertBeerFeatureQuery string

	//go:embed sql/get_beers.sql
	getBeersQuery string
)

// BeerRepository определяет контракт для хранения и получения данных о пиве.
type BeerRepository interface {

	// InsertBeer сохраняет новую сущность Beer в хранилище.
	InsertBeer(ctx context.Context, beer entities.Beer) (int, error)

	// GetBeers возвращает список всех сортов пива.
	GetBeers(ctx context.Context) ([]entities.Beer, error)
}

// BeerPostgres хранит в себе пул подлючений к БД
type BeerPostgres struct {
	Pool *pgxpool.Pool
}

// NewBeerPostgres создает новый репозиторий БД
func NewBeerPostgres(pgPool *pgxpool.Pool) *BeerPostgres {
	return &BeerPostgres{Pool: pgPool}
}

func (r *BeerPostgres) InsertBeer(ctx context.Context, beer entities.Beer) (int, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
			_ = fmt.Errorf("InsertBeer: rollback error: %w", rollbackErr)
		}
	}(tx, ctx)

	var countryID int

	err = tx.QueryRow(ctx, getOrCreateCountryQuery, beer.Country).Scan(&countryID)
	if err != nil {
		return 0, fmt.Errorf("country QueryRow: %w", err)
	}

	var cityID int

	err = tx.QueryRow(ctx, getOrCreateCityQuery, beer.City, countryID).Scan(&cityID)
	if err != nil {
		return 0, fmt.Errorf("city QueryRow: %w", err)
	}

	var typeID int

	err = tx.QueryRow(ctx, getOrCreateTypeQuery, beer.Type).Scan(&typeID)
	if err != nil {
		return 0, fmt.Errorf("type QueryRow: %w", err)
	}

	var categoryID int

	err = tx.QueryRow(ctx, getProductCategoryByNameQuery, beer.Category.Name).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("category QueryRow: %w", err)
	}

	var beerID int

	err = tx.QueryRow(ctx, insertBeerQuery,
		beer.Name, beer.Rating, beer.Description,
		beer.ABV, beer.IBU, typeID, cityID, categoryID).
		Scan(&beerID)
	if err != nil {
		return 0, fmt.Errorf("beer QueryRow: %w", err)
	}

	for _, featName := range beer.Features {
		var featID int

		err = tx.QueryRow(ctx, insertFeatureQuery, featName).Scan(&featID)
		if err != nil {
			return 0, fmt.Errorf("feature QueryRow: %w", err)
		}

		_, err = tx.Exec(ctx, insertBeerFeatureQuery, beerID, featID)
		if err != nil {
			return 0, fmt.Errorf("exec: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return beerID, nil
}

// TODO: Добавить ошибки приложения

func (r *BeerPostgres) GetBeers(ctx context.Context) ([]entities.Beer, error) {
	rows, err := r.Pool.Query(ctx, getBeersQuery)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	beers := make([]entities.Beer, 0)

	for rows.Next() {
		b := entities.Beer{}

		err = rows.Scan(&b.ID, &b.Name, &b.Rating, &b.Description,
			&b.ABV, &b.IBU, &b.City, &b.Country,
			&b.Category.Name, &b.Type, &b.Features)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		beers = append(beers, b)
	}

	return beers, nil
}

// func (r *Repository) GetBeerByName(ctx context.Context, name string) error {

// 	beers := sq.Select("*").From("beers")
// 	active := beers.Where(sq.Eq{"name": name})
// 	sql, args, err := active.ToSql()
// 	if err != nil {
// 		return fmt.Errorf("ToSql: %w", err)
// 	}
// 	row = r.Pool.QueryRow(ctx, sql, args...)

// 	return nil
// }
