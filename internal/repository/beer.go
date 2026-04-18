// Package repository содержит слой для манипуляции объектами в базе данных
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

// BeerRepository определяет контракт для хранения и получения данных о пиве.
type BeerRepository interface {

	// InsertBeer сохраняет новую сущность Beer в хранилище.
	InsertBeer(ctx context.Context, beer entities.Beer) (uint, error)

	// GetBeers возвращает список всех сортов пива.
	GetBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error)

	// UpdateBeer обновляет поля у сущности Beer в хранилище
	UpdateBeer(ctx context.Context, id uint, updates map[string]any) (uint, error)

	// DeleteBeer удаляет сущность Beer из хранилища
	DeleteBeer(ctx context.Context, id uint) error

	// InsertReview сохраняет новую сущность Review в хранилище.
	InsertReview(ctx context.Context, review entities.Review) (uint, error)

	// GetBeersByCategoryID возвращает список сортов пива, принадлежащих к определенной категории.
	GetBeersByCategoryID(ctx context.Context, ctgID uint, limit, offset uint64) ([]entities.Beer, error)

	// GetCountryID возвращает ID страны по ее названию. Если страны нет, она будет добавлена в базу данных.
	GetCountryID(ctx context.Context, name string) (uint, error)

	// GetCityID возвращает ID города по его названию и ID страны. Если города нет, он будет добавлен в базу данных.
	GetCityID(ctx context.Context, cityName string, countryID uint) (uint, error)

	// GetFeatureID возвращает ID характеристики по ее названию. Если характеристики нет, она будет добавлена в базу данных.
	GetFeatureID(ctx context.Context, featName string) (uint, error)

	// InsertBeerFeature связывает характеристику с сортом пива. Если связь уже существует, она не будет добавлена повторно.
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

// InsertBeer сохраняет новую сущность Beer в хранилище. Если страна, город, категория или характеристика не существуют, они будут добавлены в базу данных.
func (r *BeerPostgres) InsertBeer(ctx context.Context, beer entities.Beer) (uint, error) {
	if r.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
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

	ctgRepo := NewCategoryPostgres(r.Pool)
	categoryID, err := ctgRepo.GetCategoryID(ctx, beer.Category.Name)
	if err != nil {
		return 0, fmt.Errorf("getCategoryID: %w", err)
	}

	if categoryID == 0 {
		categoryID, err = ctgRepo.InsertCategory(ctx, beer.Category)
		if err != nil {
			return 0, fmt.Errorf("insertCategory: %w", err)
		}
	}

	var beerID uint
	psql := queries.InsertBeer(beer, cityID, categoryID)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("ToSql: %w", err)
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&beerID)
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

// GetBeers возвращает список всех сортов пива. Если limit не равен 0, возвращает не более limit сортов пива, начиная с позиции offset.
func (r *BeerPostgres) GetBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.FullBeerSelect().Offset(offset)
	if limit != 0 {
		psql = psql.Limit(limit)
	}

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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return beers, nil
}

// GetBeerByID возвращает сорт пива по его ID. Если сорт пива с таким ID не найден, возвращает ошибку.
func (r *BeerPostgres) GetBeerByID(ctx context.Context, id uint) (*entities.Beer, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectBeerByID(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	beer, err := scanBeer(r.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return beer, nil
}

// UpdateBeer обновляет поля у сущности Beer в хранилище. Если сорт пива с таким ID не найден, возвращает ошибку. Если updates пустой, возвращает ID без изменений.
func (r *BeerPostgres) UpdateBeer(ctx context.Context, id uint, updates map[string]any) (uint, error) {
	if r.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
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

// DeleteBeer удаляет сущность Beer из хранилища. Если сорт пива с таким ID не найден, возвращает ошибку.
func (r *BeerPostgres) DeleteBeer(ctx context.Context, id uint) error {
	if r.Pool == nil {
		return errors.New("pool is nil")
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", "Begin", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		rollbackErr := tx.Rollback(ctx)
		log, ok := logger.GetLoggerFromCtx(ctx)
		if ok {
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
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

// InsertReview сохраняет новую сущность Review в хранилище. Если сорт пива, к которому относится отзыв, не найден, возвращает ошибку.
func (r *BeerPostgres) InsertReview(ctx context.Context, review entities.Review) (uint, error) {
	if r.Pool == nil {
		return 0, errors.New("pool is nil")
	}

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

	return reviewID, nil
}

// GetBeersByCategoryID возвращает список сортов пива, принадлежащих к определенной категории. Если категория с таким ID не найдена, возвращает пустой список.
func (r *BeerPostgres) GetBeersByCategoryID(
	ctx context.Context, ctgID uint, limit, offset uint64,
) ([]entities.Beer, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectBeerByCategoryID(ctgID).Offset(offset)
	if limit != 0 {
		psql = psql.Limit(limit)
	}
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}
	rows, err := r.Pool.Query(ctx, query, args...)
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return beers, nil
}

// GetCountryID возвращает ID страны по ее названию. Если страны нет, она будет добавлена в базу данных. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) GetCountryID(ctx context.Context, name string) (uint, error) {
	if r.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	if name == "" {
		return 0, errors.New("country name is empty")
	}

	var countryID uint
	psql := queries.SelectOrInsertCountry(name)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = r.Pool.QueryRow(ctx, query, args...).Scan(&countryID)
	if err != nil {
		return 0, fmt.Errorf("country QueryRow: %w", err)
	}

	return countryID, nil
}

// GetCityID возвращает ID города по его названию и ID страны. Если города нет, он будет добавлен в базу данных. Если cityName пустой, возвращает ошибку.
func (r *BeerPostgres) GetCityID(ctx context.Context, cityName string, countryID uint) (uint, error) {
	if r.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	if cityName == "" {
		return 0, errors.New("city name is empty")
	}

	var cityID uint
	psql := queries.SelectOrInsertCity(cityName, countryID)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = r.Pool.QueryRow(ctx, query, args...).Scan(&cityID)
	if err != nil {
		return 0, fmt.Errorf("city QueryRow: %w", err)
	}

	return cityID, nil
}

// GetFeatureID возвращает ID характеристики по ее названию. Если характеристики нет, она будет добавлена в базу данных. Если featName пустой, возвращает ошибку.
func (r *BeerPostgres) GetFeatureID(ctx context.Context, featName string) (uint, error) {
	if r.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	var featID uint
	psql := queries.SelectOrInsertFeature(featName)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = r.Pool.QueryRow(ctx, query, args...).Scan(&featID)
	if err != nil {
		return 0, fmt.Errorf("category QueryRow: %w", err)
	}

	return featID, nil
}

// InsertBeerFeature связывает характеристику с сортом пива. Если связь уже существует, она не будет добавлена повторно. Если featID или beerID равны 0, возвращает ошибку.
func (r *BeerPostgres) InsertBeerFeature(ctx context.Context, featID, beerID uint) error {
	if r.Pool == nil {
		return errors.New("pool is nil")
	}

	psql := queries.SelectOrInsertBeerFeature(featID, beerID)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	_, err = r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("category QueryRow: %w", err)
	}

	return nil
}

// scanBeer сканирует строку из базы данных в сущность Beer. Если строка не соответствует структуре сущности, возвращает ошибку.
func scanBeer(row pgx.Row) (*entities.Beer, error) {
	var beer entities.Beer
	err := row.Scan(&beer.ID, &beer.Name, &beer.Rating,
		&beer.Description, &beer.ABV, &beer.IBU, &beer.Amount,
		&beer.Unit, &beer.City, &beer.Country,
		&beer.Category.Name, &beer.Features)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	return &beer, nil
}
