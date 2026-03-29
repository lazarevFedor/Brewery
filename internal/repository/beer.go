// Package repository contains layer that manipulates data in database
package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"Brewery/pkg/logger"
	"context"
	"errors"
	"fmt"

	_ "embed"

	"go.uber.org/zap"

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

	//go:embed sql/get_or_create_category.sql
	getOrCreateCategoryQuery string

	//go:embed sql/insert_beer.sql
	insertBeerQuery string

	//go:embed sql/get_or_create_feature.sql
	getOrCreateFeatureQuery string

	//go:embed sql/insert_beer_feature.sql
	insertBeerFeatureQuery string
)

// BeerRepository определяет контракт для хранения и получения данных о пиве.
type BeerRepository interface {

	// InsertBeer сохраняет новую сущность Beer в хранилище.
	InsertBeer(ctx context.Context, beer entities.Beer) (uint, error)

	// GetBeers возвращает список всех сортов пива.
	GetBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error)

	UpdateBeer(ctx context.Context, id uint, updates map[string]any) (uint, error)

	DeleteBeer(ctx context.Context, id uint) error

	InsertReview(ctx context.Context, review entities.Review) (uint, error)

	GetBeersByCategoryID(ctx context.Context, ctgID uint, limit, offset uint64) ([]entities.Beer, error)

	GetCountryID(ctx context.Context, name string) (uint, error)

	GetCityID(ctx context.Context, cityName string, countryID uint) (uint, error)

	GetTypeID(ctx context.Context, typeName string) (uint, error)

	GetFeatureID(ctx context.Context, featName string) (uint, error)

	InsertBeerFeature(ctx context.Context, featID, beerID uint) error
}

// BeerPostgres хранит в себе пул подключений к БД
type BeerPostgres struct {
	Pool *pgxpool.Pool
}

// NewBeerPostgres создает новый репозиторий БД
func NewBeerPostgres(pgPool *pgxpool.Pool) *BeerPostgres {
	return &BeerPostgres{Pool: pgPool}
}

func (r *BeerPostgres) InsertBeer(ctx context.Context, beer entities.Beer) (uint, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	countryID, err := r.GetCountryID(ctx, beer.Country)
	if err != nil {
		return 0, fmt.Errorf("country QueryRow: %w", err)
	}

	cityID, err := r.GetCityID(ctx, beer.City, countryID)
	if err != nil {
		return 0, fmt.Errorf("city QueryRow: %w", err)
	}

	typeID, err := r.GetTypeID(ctx, beer.Category.Name)
	if err != nil {
		return 0, fmt.Errorf("type QueryRow: %w", err)
	}

	ctgRepo := NewCategoryPostgres(r.Pool)
	categoryID, err := ctgRepo.GetCategoryID(ctx, beer.Category.Name)
	if err != nil {
		return 0, fmt.Errorf("category QueryRow: %w", err)
	}

	var beerID uint
	err = tx.QueryRow(ctx, insertBeerQuery,
		beer.Name, beer.Rating, beer.Description,
		beer.ABV, beer.IBU, typeID, cityID, categoryID).
		Scan(&beerID)
	if err != nil {
		return 0, fmt.Errorf("beer QueryRow: %w", err)
	}

	for _, featName := range beer.Features {
		featID, err := r.GetFeatureID(ctx, featName)
		if err != nil {
			return 0, fmt.Errorf("feature QueryRow: %w", err)
		}

		err = r.InsertBeerFeature(ctx, featID, beerID)
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
func (r *BeerPostgres) GetBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.FullBeerSelect().Offset(offset)
	if limit != 0 {
		psql = psql.Limit(limit)
	}

	query, _, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	beers := make([]entities.Beer, 0)
	for rows.Next() {
		beer, err := scanBeer(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		beers = append(beers, *beer)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return beers, nil
}

func (r *BeerPostgres) GetBeerByID(ctx context.Context, id uint) (*entities.Beer, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.SelectBeerByID(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	beer, err := scanBeer(tx.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return beer, nil
}

func (r *BeerPostgres) UpdateBeer(ctx context.Context, id uint, updates map[string]any) (uint, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.UpdateBeer(id, updates)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var beerID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&beerID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Scan", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return beerID, nil
}

func (r *BeerPostgres) DeleteBeer(ctx context.Context, id uint) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.DeleteBeer(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}
	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "Exec", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("failed to delete beer")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *BeerPostgres) InsertReview(ctx context.Context, review entities.Review) (uint, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.InsertReview(review)

	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var reviewID uint
	err = r.Pool.QueryRow(ctx, query, args...).Scan(&reviewID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Scan", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return reviewID, nil
}

func (r *BeerPostgres) GetBeersByCategoryID(
	ctx context.Context, ctgID uint, limit, offset uint64,
) ([]entities.Beer, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && errors.Is(rollbackErr, pgx.ErrTxClosed) {
				log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
			}
		}
	}(tx, ctx)

	psql := queries.SelectBeerByCategoryID(ctgID).Offset(offset)
	if limit != 0 {
		psql = psql.Limit(limit)
	}
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	beers := make([]entities.Beer, 0)
	for rows.Next() {
		beer, err := scanBeer(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		beers = append(beers, *beer)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return beers, nil
}

func (r *BeerPostgres) GetCountryID(ctx context.Context, name string) (uint, error) {
	var countryID uint
	err := r.Pool.QueryRow(ctx, getOrCreateCountryQuery, name).Scan(&countryID)
	if err != nil {
		return 0, fmt.Errorf("country QueryRow: %w", err)
	}

	return countryID, nil
}

func (r *BeerPostgres) GetCityID(ctx context.Context, cityName string, countryID uint) (uint, error) {
	var cityID uint
	err := r.Pool.QueryRow(ctx, getOrCreateCityQuery, cityName, countryID).Scan(&cityID)
	if err != nil {
		return 0, fmt.Errorf("city QueryRow: %w", err)
	}

	return cityID, nil
}

func (r *BeerPostgres) GetTypeID(ctx context.Context, typeName string) (uint, error) {
	var typeID uint
	err := r.Pool.QueryRow(ctx, getOrCreateTypeQuery, typeName).Scan(&typeID)
	if err != nil {
		return 0, fmt.Errorf("type QueryRow: %w", err)
	}

	return typeID, nil
}

func (r *BeerPostgres) GetFeatureID(ctx context.Context, featName string) (uint, error) {
	var featID uint
	err := r.Pool.QueryRow(ctx, getOrCreateFeatureQuery, featName).Scan(&featID)
	if err != nil {
		return 0, fmt.Errorf("category QueryRow: %w", err)
	}

	return featID, nil
}

func (r *BeerPostgres) InsertBeerFeature(ctx context.Context, featID, beerID uint) error {
	_, err := r.Pool.Exec(ctx, insertBeerFeatureQuery, beerID, featID)
	if err != nil {
		return fmt.Errorf("category QueryRow: %w", err)
	}

	return nil
}

func scanBeer(row pgx.Row) (*entities.Beer, error) {
	var beer entities.Beer
	err := row.Scan(&beer.ID, &beer.Name, &beer.Rating, &beer.Description,
		&beer.ABV, &beer.IBU, &beer.City, &beer.Country,
		&beer.Category.Name, &beer.Type, &beer.Features)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	return &beer, nil
}
