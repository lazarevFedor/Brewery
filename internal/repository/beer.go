// Package repository contains layer that manipulates data in database
package repository

import (
	"Brewery/internal/entities"
	"Brewery/pkg/logger"
	"context"
	"errors"
	"fmt"

	_ "embed"

	sq "github.com/Masterminds/squirrel"
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

	//go:embed sql/get_product_category_by_name.sql
	getProductCategoryByNameQuery string

	//go:embed sql/insert_beer.sql
	insertBeerQuery string

	//go:embed sql/insert_feature.sql
	insertFeatureQuery string

	//go:embed sql/insert_beer_feature.sql
	insertBeerFeatureQuery string
)

// BeerRepository определяет контракт для хранения и получения данных о пиве.
type BeerRepository interface {

	// InsertBeer сохраняет новую сущность Beer в хранилище.
	InsertBeer(ctx context.Context, beer entities.Beer) (int, error)

	// GetBeers возвращает список всех сортов пива.
	GetBeers(ctx context.Context, limit, offset int) ([]entities.Beer, error)

	UpdateBeer(ctx context.Context, id int, updates map[string]any) (*entities.Beer, error)

	DeleteBeer(ctx context.Context, id int) error

	InsertReview(ctx context.Context, review entities.Review) (int, error)

	GetBeersByCategoryID(ctx context.Context, ctgID, limit, offset int) ([]entities.Beer, error)
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
func (r *BeerPostgres) GetBeers(ctx context.Context, limit, offset int) ([]entities.Beer, error) {
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

	builder := sq.Select("*").From("beers").
		OrderBy("id DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, _, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}
	rows, err := r.Pool.Query(ctx, query)
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

func (r *BeerPostgres) UpdateBeer(ctx context.Context, id int, updates map[string]any) (*entities.Beer, error) {
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
	builder := sq.Update("beers").
		SetMap(updates).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *")
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	beer, err := scanBeer(r.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return beer, nil
}

func (r *BeerPostgres) DeleteBeer(ctx context.Context, id int) error {
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
	condition := map[string]any{
		"id": id,
	}
	builder := sq.Delete("beers").Where(condition)
	psql := builder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}
	result, err := r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "Exec", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("failed to delete category: no such category")
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *BeerPostgres) InsertReview(ctx context.Context, review entities.Review) (int, error) {
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
	data := map[string]any{
		"body":    review.Body,
		"rating":  review.Rating,
		"beer_id": review.BeerID,
	}

	builder := sq.Insert("reviews").SetMap(data).Suffix("RETURNING id")
	psql := builder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var reviewID int
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

func (r *BeerPostgres) GetBeersByCategoryID(ctx context.Context, ctgID, limit, offset int) ([]entities.Beer, error) {
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

	builder := sq.Select("*").From("beers").
		Where(sq.Eq{"category_id": ctgID}).
		OrderBy("id DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
	psql := builder.PlaceholderFormat(sq.Dollar)
	query, _, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}
	rows, err := r.Pool.Query(ctx, query)
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
