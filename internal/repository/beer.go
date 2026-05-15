// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"Brewery/pkg/logger"
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BeerRepository определяет контракт для хранения и получения данных о пиве.
type BeerRepository interface {

	// BeerExists проверяет наличие объекта по id.
	BeerExists(ctx context.Context, id uint) (bool, error)

	// InsertBeer сохраняет новую сущность Beer в хранилище.
	InsertBeer(ctx context.Context, beer entities.Beer) (*entities.Beer, error)

	// GetBeers возвращает список всех сортов пива.
	GetBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error)

	// UpdateBeer обновляет поля у сущности Beer в хранилище
	UpdateBeer(ctx context.Context, id uint, updates map[string]any) (*entities.Beer, error)

	// DeleteBeer удаляет сущность Beer из хранилища
	DeleteBeer(ctx context.Context, id uint) error

	// FilterBeer возвращает список сущностей пиво по фильтру
	FilterBeer(ctx context.Context, filters []*entities.FilterParameter, limit, offset uint64, categoryID uint) ([]entities.Beer, error)

	// InsertReview сохраняет новую сущность Review в хранилище.
	InsertReview(ctx context.Context, review entities.Review) (uint, error)

	// DeleteReview удаляет сущность Review из хранилища
	DeleteReview(ctx context.Context, id uint) error

	// UpdateReview обновляет поля у сущности Review в хранилище.
	UpdateReview(ctx context.Context, id uint, updates map[string]any) error

	// GetReviews возвращает список всех отзывов конкретного пива.
	GetReviews(ctx context.Context, limit, offset uint64, beerID uint) ([]entities.Review, error)

	// GetBeersByCategoryID возвращает список сортов пива, принадлежащих к определенной категории.
	GetBeersByCategoryID(ctx context.Context, ctgID uint, limit, offset uint64) ([]entities.Beer, error)

	// GetCountryID возвращает ID страны по ее названию. Если страны нет, она будет добавлена в базу данных.
	GetCountryID(ctx context.Context, name string) (uint, error)

	// GetCityID возвращает ID города по его названию и ID страны. Если города нет, он будет добавлен в базу данных.
	GetCityID(ctx context.Context, cityName string, countryID uint) (uint, error)

	// GetCityNameByID возвращает название города по его ID.
	GetCityNameByID(ctx context.Context, id uint) (string, error)

	// GetFeatureID возвращает ID характеристики по ее названию. Если характеристики нет, она будет добавлена в базу данных.
	GetFeatureID(ctx context.Context, featName string) (uint, error)

	// GetBeerFeature и тд
	GetBeerFeature(ctx context.Context, beerID uint) ([]string, error)

	// InsertBeerFeature связывает характеристику с сортом пива. Если связь уже существует, она не будет добавлена повторно.
	InsertBeerFeature(ctx context.Context, featID, beerID uint) error
}

// BeerPostgres хранит в себе пул подключений к БД
type BeerPostgres struct {
	Pool *pgxpool.Pool
}

// beersBufferCapacity определяет начальную емкость среза для хранения сортов пива при чтении из базы данных.
const beersBufferCapacity = 10

// beerSlicePool пул слайсов для хранения сортов пива.
var beerSlicePool = sync.Pool{
	New: func() any { return new(make([]entities.Beer, 0, beersBufferCapacity)) },
}

// NewBeerRepository создает новый экземпляр BeerRepository с переданным пулом соединений.
func NewBeerRepository(pgPool *pgxpool.Pool) BeerRepository {
	return &BeerPostgres{Pool: pgPool}
}

// NewBeerPostgres создает новый репозиторий БД
func NewBeerPostgres(pgPool *pgxpool.Pool) *BeerPostgres {
	return &BeerPostgres{Pool: pgPool}
}

func (r *BeerPostgres) BeerExists(ctx context.Context, id uint) (bool, error) {
	if r.Pool == nil {
		return false, errors.New("pool is nil")
	}

	psql := queries.Exists(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return false, fmt.Errorf("ToSql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return false, err
	}

	vals, err := rows.Values()
	if err != nil {
		return false, err
	}

	if len(vals) == 0 {
		return false, nil
	}

	return true, nil
}

// InsertBeer сохраняет новую сущность Beer в хранилище. Если страна, город, категория или характеристика не существуют, они будут добавлены в базу данных.
func (r *BeerPostgres) InsertBeer(ctx context.Context, beer entities.Beer) (*entities.Beer, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Begin", err)
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

	countryID, err := r.getCountryIDTx(ctx, tx, beer.Country)
	if err != nil {
		return nil, fmt.Errorf("country QueryRow: %w", err)
	}

	cityID, err := r.getCityIDTx(ctx, tx, beer.City, countryID)
	if err != nil {
		return nil, fmt.Errorf("city QueryRow: %w", err)
	}

	categoryID, err := r.getCategoryIDTx(ctx, tx, beer.Category.Name)
	if err != nil {
		return nil, fmt.Errorf("getCategoryID: %w", err)
	}

	if categoryID == 0 {
		categoryID, err = r.insertCategoryTx(ctx, tx, beer.Category)
		if err != nil {
			return nil, fmt.Errorf("insertCategory: %w", err)
		}
	}

	psql := queries.InsertBeer(beer, cityID, categoryID)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	row := tx.QueryRow(ctx, query, args...)
	createdBeer, err := scanBeerBase(row)

	if err != nil {
		return nil, fmt.Errorf("beer QueryRow: %w", err)
	}

	createdBeer.City = beer.City
	createdBeer.Country = beer.Country
	createdBeer.Category.Name = beer.Category.Name

	for _, featName := range beer.Features {
		featID, err := r.getFeatureIDTx(ctx, tx, featName)
		if err != nil {
			return nil, fmt.Errorf("feature QueryRow: %w", err)
		}

		err = r.insertBeerFeatureTx(ctx, tx, featID, createdBeer.ID)
		if err != nil {
			return nil, fmt.Errorf("exec: %w", err)
		}
	}
	createdBeer.Features = beer.Features

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return createdBeer, nil
}

// getCountryIDTx возвращает ID страны по ее названию в рамках транзакции. Если страны нет, она будет добавлена в базу данных. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) getCountryIDTx(ctx context.Context, tx pgx.Tx, name string) (uint, error) {
	if name == "" {
		return 0, errors.New("country name is empty")
	}

	var countryID uint
	psql := queries.SelectOrInsertCountry(name)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&countryID)
	if err != nil {
		return 0, fmt.Errorf("country QueryRow: %w", err)
	}

	return countryID, nil
}

// getCityIDTx
func (r *BeerPostgres) getCityIDTx(ctx context.Context, tx pgx.Tx, name string, countryID uint) (uint, error) {
	if name == "" {
		return 0, errors.New("city name is empty")
	}

	var cityID uint
	psql := queries.SelectOrInsertCity(name, countryID)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&cityID)
	if err != nil {
		return 0, fmt.Errorf("city QueryRow: %w", err)
	}

	return cityID, nil
}

// GetCityNameByID возвращает название города по его ID.
// TODO: перенести в queries
func (r *BeerPostgres) GetCityNameByID(ctx context.Context, id uint) (string, error) {
	if r.Pool == nil {
		return "", errors.New("pool is nil")
	}

	var name string
	query := "SELECT name FROM cities WHERE id = $1"
	if err := r.Pool.QueryRow(ctx, query, id).Scan(&name); err != nil {
		return "", fmt.Errorf("city QueryRow: %w", err)
	}
	return name, nil
}

// getCategoryIDTx возвращает ID категории по ее названию в рамках транзакции. Если категории нет, возвращает 0. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) getCategoryIDTx(ctx context.Context, tx pgx.Tx, categoryName string) (uint, error) {
	if categoryName == "" {
		return 0, errors.New("category name cannot be empty")
	}

	var categoryID uint
	psql := queries.SelectCategoryByName(categoryName)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&categoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		return 0, fmt.Errorf("scan: %w", err)
	}

	return categoryID, nil
}

// insertCategoryTx вставляет категорию в базу данных в рамках транзакции и возвращает ее ID. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) insertCategoryTx(ctx context.Context, tx pgx.Tx, category entities.ProductCategory) (uint, error) {
	psql := queries.CategoryInsert(category)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var categoryID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("exec: %w", err)
	}

	if categoryID == 0 {
		return 0, errors.New("zero id")
	}

	return categoryID, nil
}

// getFeatureIDTx возвращает ID характеристики по ее названию в рамках транзакции. Если характеристики нет, она будет добавлена в базу данных. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) getFeatureIDTx(ctx context.Context, tx pgx.Tx, name string) (uint, error) {
	var featID uint
	psql := queries.SelectOrInsertFeature(name)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&featID)
	if err != nil {
		return 0, fmt.Errorf("QueryRow: %w", err)
	}

	return featID, nil
}

// insertBeerFeatureTx связывает характеристику с сортом пива в рамках транзакции. Если связь уже существует, она не будет добавлена повторно. Если featID или beerID равны 0, возвращает ошибку.
func (r *BeerPostgres) insertBeerFeatureTx(ctx context.Context, tx pgx.Tx, featID, beerID uint) error {
	psql := queries.SelectOrInsertBeerFeature(featID, beerID)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("QueryRow: %w", err)
	}

	return nil
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

	bufp := beerSlicePool.Get().(*[]entities.Beer)
	buf := *bufp
	for rows.Next() {
		beer, err := scanBeer(rows)
		if err != nil {
			clear(buf)
			*bufp = buf[:0]
			beerSlicePool.Put(bufp)
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		buf = append(buf, *beer)
	}

	if err = rows.Err(); err != nil {
		clear(buf)
		*bufp = buf[:0]
		beerSlicePool.Put(bufp)
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	beers := make([]entities.Beer, len(buf))
	copy(beers, buf)
	clear(buf)
	*bufp = buf[:0]
	beerSlicePool.Put(bufp)

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
func (r *BeerPostgres) UpdateBeer(ctx context.Context, id uint, updates map[string]any) (*entities.Beer, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.UpdateBeer(id, updates)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	row := r.Pool.QueryRow(ctx, query, args...)
	beer, err := scanBeerBase(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	return beer, nil
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

func (r *BeerPostgres) FilterBeer(ctx context.Context, filters []*entities.FilterParameter, limit, offset uint64, categoryID uint) ([]entities.Beer, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.FilterBeers(filters, categoryID).Offset(offset)
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

	bufp := beerSlicePool.Get().(*[]entities.Beer)
	buf := *bufp
	for rows.Next() {
		beer, err := scanBeer(rows)
		if err != nil {
			clear(buf)
			*bufp = buf[:0]
			beerSlicePool.Put(bufp)
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		buf = append(buf, *beer)
	}

	if err = rows.Err(); err != nil {
		clear(buf)
		*bufp = buf[:0]
		beerSlicePool.Put(bufp)
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	beers := make([]entities.Beer, len(buf))
	copy(beers, buf)
	clear(buf)
	*bufp = buf[:0]
	beerSlicePool.Put(bufp)

	return beers, nil
}

// InsertReview сохраняет новую сущность Review в хранилище. Если сорт пива, к которому относится отзыв, не найден, возвращает ошибку.
func (r *BeerPostgres) InsertReview(ctx context.Context, review entities.Review) (uint, error) {
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

	insertPsql := queries.InsertReview(review)

	query, args, err := insertPsql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var reviewID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&reviewID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Scan", err)
	}

	updatePsql := queries.UpdateBeerRating(review.BeerID, review.Rating, "insert")
	query, args, err = updatePsql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "Scan", err)
	}

	if result.RowsAffected() != 1 {
		return 0, fmt.Errorf("%s: %w", "RowsAffected", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return reviewID, nil
}

// DeleteReview удаляет сущность Review из хранилища. Если отзыв с таким id, не найден, возвращает ошибку.
func (r *BeerPostgres) DeleteReview(ctx context.Context, id uint) error {
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

	psql := queries.DeleteReview(id)

	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	var rating, beerID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&beerID, &rating)
	if err != nil {
		return fmt.Errorf("%s: %w", "Exec", err)
	}

	updatePsql := queries.UpdateBeerRating(beerID, rating, "delete")
	query, args, err = updatePsql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "Scan", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("%s: %w, %d, %d", "RowsAffected", err, rating, beerID)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// UpdateReview обновляет поля у сущности Review в хранилище. Если отзыв с таким id, не найден, возвращает ошибку.
func (r *BeerPostgres) UpdateReview(ctx context.Context, id uint, updates map[string]any) error {
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

	psql := queries.UpdateReview(id, updates)

	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	var rating, beerID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&beerID, &rating)
	if err != nil {
		return fmt.Errorf("%s: %w", "QueryRow", err)
	}

	updatePsql := queries.UpdateBeerRating(beerID, rating, "delete")
	query, args, err = updatePsql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "Scan", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("%s: %w", "RowsAffected", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// GetReviews возвращает список всех отзывов к конкретному пиву, возвращает не более limit отзывов, начиная с позиции offset.
func (r *BeerPostgres) GetReviews(ctx context.Context, limit, offset uint64, beerID uint) ([]entities.Review, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectReviewByBeerID(beerID).Offset(offset)
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

	reviews := make([]entities.Review, 0)
	for rows.Next() {
		var review entities.Review

		err = rows.Scan(&review.ID, &review.Body, &review.BeerID, &review.Rating)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return reviews, nil
}

// GetBeersByCategoryID возвращает список сортов пива, принадлежащих к определенной категории. Если категория с таким ID не найдена, возвращает пустой список.
func (r *BeerPostgres) GetBeersByCategoryID(ctx context.Context, ctgID uint, limit, offset uint64) ([]entities.Beer, error) {
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

	bufp := beerSlicePool.Get().(*[]entities.Beer)
	buf := *bufp
	for rows.Next() {
		beer, err := scanBeer(rows)
		if err != nil {
			clear(buf)
			*bufp = buf[:0]
			beerSlicePool.Put(bufp)
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		buf = append(buf, *beer)
	}

	if err = rows.Err(); err != nil {
		clear(buf)
		*bufp = buf[:0]
		beerSlicePool.Put(bufp)
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	beers := make([]entities.Beer, len(buf))
	copy(beers, buf)
	clear(buf)
	*bufp = buf[:0]
	beerSlicePool.Put(bufp)

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

// GetCityID возвращает ID города по его названию и ID страны. Если города нет, он будет добавлен в базу данных. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) GetCityID(ctx context.Context, name string, countryID uint) (uint, error) {
	if r.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	if name == "" {
		return 0, errors.New("city name is empty")
	}

	var cityID uint
	psql := queries.SelectOrInsertCity(name, countryID)
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

// GetFeatureID возвращает ID характеристики по ее названию. Если характеристики нет, она будет добавлена в базу данных. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) GetFeatureID(ctx context.Context, name string) (uint, error) {
	if r.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	var featID uint
	psql := queries.SelectOrInsertFeature(name)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = r.Pool.QueryRow(ctx, query, args...).Scan(&featID)
	if err != nil {
		return 0, fmt.Errorf("QueryRow: %w", err)
	}

	return featID, nil
}

// GetBeerFeature возвращает список и тд
func (r *BeerPostgres) GetBeerFeature(ctx context.Context, beerID uint) ([]string, error) {
	if r.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectBeersFeature(beerID)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("QueryRow: %w", err)
	}
	defer rows.Close()

	features := make([]string, 0)
	var featName string
	for rows.Next() {
		err := rows.Scan(&featName)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		features = append(features, featName)
	}

	return features, nil
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
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

// scanBeer сканирует обработанную строку из базы данных в сущность Beer. Если строка не соответствует структуре сущности, возвращает ошибку.
func scanBeer(row pgx.Row) (*entities.Beer, error) {
	var beer entities.Beer
	var reviewRatingSum, reviewAmount uint
	err := row.Scan(&beer.ID, &beer.Name,
		&beer.Description, &beer.ABV, &beer.IBU, &beer.Amount,
		&beer.Unit, &beer.City, &beer.Country,
		&beer.Category.Name, &beer.Features,
		&reviewRatingSum, &reviewAmount)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	if reviewAmount != 0 {
		rating := float32(reviewRatingSum) / float32(reviewAmount)
		beer.Rating = rating
	}

	return &beer, nil
}

// scanBeerBase сканирует полную сырую строку из базы данных в сущность Beer. Если строка не соответствует структуре сущности, возвращает ошибку.
func scanBeerBase(row pgx.Row) (*entities.Beer, error) {
	var beer entities.Beer
	var cityID, categoryID, reviewRatingSum, reviewAmount uint
	err := row.Scan(&beer.ID, &beer.Name,
		&beer.Description, &beer.ABV, &beer.IBU,
		&beer.Amount, &beer.Unit, &cityID, &categoryID,
		&reviewRatingSum, &reviewAmount)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	if reviewAmount != 0 {
		rating := float32(reviewRatingSum) / float32(reviewAmount)
		beer.Rating = rating
	}

	return &beer, nil
}
