// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/apperrors"
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
	GetCountryID(ctx context.Context, tx pgx.Tx, name string) (uint, error)

	// GetCityID возвращает ID города по его названию и ID страны. Если города нет, он будет добавлен в базу данных.
	GetCityID(ctx context.Context, tx pgx.Tx, cityName string, countryID uint) (uint, error)

	// GetCityNameByID возвращает название города по его ID.
	GetCityNameByID(ctx context.Context, id uint) (string, error)

	// GetFeatureID возвращает ID характеристики по ее названию. Если характеристики нет, она будет добавлена в базу данных.
	GetFeatureID(ctx context.Context, tx pgx.Tx, featName string) (uint, error)

	// GetBeerFeature и тд
	GetBeerFeature(ctx context.Context, beerID uint) ([]string, error)

	// ConnectBeerAndFeature связывает характеристику с сортом пива. Если связь уже существует, она не будет добавлена повторно.
	ConnectBeerAndFeature(ctx context.Context, tx pgx.Tx, featID, beerID uint) error

	// DisconnectBeerAndFeature удаляет связь характеристики с сортом пива. Если связи нет, возвращает ошибку.
	DisconnectBeerAndFeature(ctx context.Context, tx pgx.Tx, beerID uint) error

	GetBeerByID(ctx context.Context, id uint) (*entities.Beer, error)
}

// BeerPostgres хранит в себе пул подключений к БД
type BeerPostgres struct {
	Pool *pgxpool.Pool
}

// beersBufferCapacity определяет начальную емкость среза для хранения сортов пива при чтении из базы данных.
const beersBufferCapacity = 10

// beerSlicePool пул слайсов для хранения сортов пива.
var beerSlicePool = sync.Pool{
	New: func() any {
		s := make([]entities.Beer, 0, beersBufferCapacity)
		return &s
	},
}

// NewBeerRepository создает новый экземпляр BeerRepository с переданным пулом соединений.
func NewBeerRepository(pgPool *pgxpool.Pool) BeerRepository {
	return &BeerPostgres{Pool: pgPool}
}

// NewBeerPostgres создает новый репозиторий БД
func NewBeerPostgres(pgPool *pgxpool.Pool) *BeerPostgres {
	return &BeerPostgres{Pool: pgPool}
}

var rollbackFunc = func(tx pgx.Tx, ctx context.Context) {
	rollbackErr := tx.Rollback(ctx)
	log, ok := logger.GetLoggerFromCtx(ctx)
	if ok {
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			log.Error(ctx, "InsertBeer: rollback error:", zap.Error(rollbackErr))
		}
	}
}

func (r *BeerPostgres) BeerExists(ctx context.Context, id uint) (bool, error) {
	if r.Pool == nil {
		return false, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.Exists()
	query, _, err := psql.ToSql()
	if err != nil {
		return false, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	var exists bool
	err = r.Pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, apperrors.Internal(fmt.Errorf("query: %w", err))
	}

	return exists, nil
}

// InsertBeer сохраняет новую сущность Beer в хранилище. Если страна, город, категория или характеристика не существуют, они будут добавлены в базу данных.
func (r *BeerPostgres) InsertBeer(ctx context.Context, beer entities.Beer) (*entities.Beer, error) {
	if r.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	ctgRepo := NewCategoryRepository(r.Pool)

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("begin: %w", err))
	}
	defer rollbackFunc(tx, ctx)

	countryID, err := r.GetCountryID(ctx, tx, beer.Country)
	if err != nil {
		return nil, err
	}

	cityID, err := r.GetCityID(ctx, tx, beer.City, countryID)
	if err != nil {
		return nil, err
	}

	categoryID, err := ctgRepo.GetCategoryID(ctx, tx, beer.Category.Name)
	if err != nil {
		return nil, err
	}

	if categoryID == 0 {
		return nil, apperrors.BadRequest("category not found", fmt.Errorf("category not found: %s", beer.Category.Name))
	}

	psql := queries.InsertBeer(beer, cityID, categoryID)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	row := tx.QueryRow(ctx, query, args...)
	createdBeer, err := scanBeerBase(row)

	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("beer QueryRow: %w", err))
	}

	createdBeer.City = beer.City
	createdBeer.Country = beer.Country
	createdBeer.Category.Name = beer.Category.Name

	for _, featName := range beer.Features {
		featID, err := r.GetFeatureID(ctx, tx, featName)
		if err != nil {
			return nil, err
		}

		err = r.ConnectBeerAndFeature(ctx, tx, featID, createdBeer.ID)
		if err != nil {
			return nil, err
		}
	}
	createdBeer.Features = beer.Features

	err = tx.Commit(ctx)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("commit: %w", err))
	}
	return createdBeer, nil
}

// GetCityNameByID возвращает название города по его ID.
func (r *BeerPostgres) GetCityNameByID(ctx context.Context, id uint) (string, error) {
	if r.Pool == nil {
		return "", errors.New("pool is nil")
	}

	var name string
	psql := queries.SelectCityNameByID(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return "", fmt.Errorf("toSql: %w", err)
	}

	if err = r.Pool.QueryRow(ctx, query, args...).Scan(&name); err != nil {
		return "", fmt.Errorf("city QueryRow: %w", err)
	}
	return name, nil
}

// GetBeers возвращает список всех сортов пива. Если limit не равен 0, возвращает не более limit сортов пива, начиная с позиции offset.
func (r *BeerPostgres) GetBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error) {
	if r.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.FullBeerSelect().Offset(offset)
	if limit != 0 {
		psql = psql.Limit(limit)
	}

	query, _, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("query: %w", err))
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
			return nil, err
		}

		buf = append(buf, *beer)
	}

	if err = rows.Err(); err != nil {
		clear(buf)
		*bufp = buf[:0]
		beerSlicePool.Put(bufp)
		return nil, apperrors.Internal(fmt.Errorf("rows.Err: %w", err))
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
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectBeerByID(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	beer, err := scanBeer(r.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("beer not found", err)
		}
		return nil, apperrors.Internal(fmt.Errorf("query: %w", err))
	}

	return beer, nil
}

// UpdateBeer обновляет поля у сущности Beer в хранилище. Если сорт пива с таким ID не найден, возвращает ошибку. Если updates пустой, возвращает ID без изменений.
func (r *BeerPostgres) UpdateBeer(ctx context.Context, id uint, updates map[string]any) (*entities.Beer, error) {
	if r.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.UpdateBeer(id, updates)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	row := r.Pool.QueryRow(ctx, query, args...)
	beer, err := scanBeerBase(row)
	if err != nil {
		return nil, err
	}

	return beer, nil
}

// DeleteBeer удаляет сущность Beer из хранилища. Если сорт пива с таким ID не найден, возвращает ошибку.
func (r *BeerPostgres) DeleteBeer(ctx context.Context, id uint) error {
	if r.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("%s: %w", "Begin", err))
	}
	defer rollbackFunc(tx, ctx)

	psql := queries.DeleteBeer(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}
	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("%s: %w", "Exec", err))
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("beer not found", nil)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("commit: %w", err))
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
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.InsertReview(review)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	var reviewID uint
	err = r.Pool.QueryRow(ctx, query, args...).Scan(&reviewID)
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("%s: %w", "scan", err))
	}

	return reviewID, nil
}

// DeleteReview удаляет сущность Review из хранилища. Если отзыв с таким id, не найден, возвращает ошибку.
func (r *BeerPostgres) DeleteReview(ctx context.Context, id uint) error {
	if r.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", "Begin", err)
	}
	defer rollbackFunc(tx, ctx)

	psql := queries.DeleteReview(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	var rating, beerID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&beerID, &rating)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("exec: %w", err))
	}

	updatePsql := queries.UpdateBeerRating(beerID, rating, "delete")
	query, args, err = updatePsql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("beer exec: %w", err))
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("review not found", nil)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("commit: %w", err))
	}
	return nil
}

// UpdateReview обновляет поля у сущности Review в хранилище. Если отзыв с таким id, не найден, возвращает ошибку.
func (r *BeerPostgres) UpdateReview(ctx context.Context, id uint, updates map[string]any) error {
	if r.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", "Begin", err)
	}
	defer rollbackFunc(tx, ctx)

	psql := queries.UpdateReview(id, updates)

	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	var rating, beerID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&beerID, &rating)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("%s: %w", "Exec", err))
	}

	updatePsql := queries.UpdateBeerRating(beerID, rating, "delete")
	query, args, err = updatePsql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("beer: exec: %w", err))
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("review not found", nil)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("commit: %w", err))
	}
	return err
}

// GetReviews возвращает список всех отзывов к конкретному пиву, возвращает не более limit отзывов, начиная с позиции offset.
func (r *BeerPostgres) GetReviews(ctx context.Context, limit, offset uint64, beerID uint) ([]entities.Review, error) {
	if r.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectReviewByBeerID(beerID).Offset(offset)
	if limit != 0 {
		psql = psql.Limit(limit)
	}

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("query: %w", err))
	}
	defer rows.Close()

	reviews := make([]entities.Review, 0)
	for rows.Next() {
		var review entities.Review

		err = rows.Scan(&review.ID, &review.Author, &review.Body, &review.BeerID, &review.Rating)
		if err != nil {
			return nil, apperrors.Internal(fmt.Errorf("scan: %w", err))
		}

		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("rows.Err: %w", err))
	}

	return reviews, nil
}

// GetBeersByCategoryID возвращает список сортов пива, принадлежащих к определенной категории. Если категория с таким ID не найдена, возвращает пустой список.
func (r *BeerPostgres) GetBeersByCategoryID(ctx context.Context, ctgID uint, limit, offset uint64) ([]entities.Beer, error) {
	if r.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectBeerByCategoryID(ctgID).Offset(offset)
	if limit != 0 {
		psql = psql.Limit(limit)
	}
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("query: %w", err))
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
			return nil, err
		}
		buf = append(buf, *beer)
	}

	if err = rows.Err(); err != nil {
		clear(buf)
		*bufp = buf[:0]
		beerSlicePool.Put(bufp)
		return nil, apperrors.Internal(fmt.Errorf("rows.Err: %w", err))
	}

	beers := make([]entities.Beer, len(buf))
	copy(beers, buf)
	clear(buf)
	*bufp = buf[:0]
	beerSlicePool.Put(bufp)

	return beers, nil
}

// GetBeerFeature возвращает список и тд
func (r *BeerPostgres) GetBeerFeature(ctx context.Context, beerID uint) ([]string, error) {
	if r.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectBeersFeature(beerID)
	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("feature QueryRow: %w", err))
	}
	defer rows.Close()

	features := make([]string, 0)
	var featName string

	for rows.Next() {
		err := rows.Scan(&featName)
		if err != nil {
			return nil, apperrors.Internal(fmt.Errorf("scan: %w", err))
		}
		features = append(features, featName)
	}

	return features, nil
}

// ConnectBeerAndFeature связывает характеристику с сортом пива в рамках транзакции. Если связь уже существует, она не будет добавлена повторно. Если featID или beerID равны 0, возвращает ошибку.
func (r *BeerPostgres) ConnectBeerAndFeature(ctx context.Context, tx pgx.Tx, featID, beerID uint) error {
	if r.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}
	psql := queries.ConnectBeerAndFeature(featID, beerID)
	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	if tx != nil {
		_, err = tx.Exec(ctx, query, args...)
	} else {
		_, err = r.Pool.Exec(ctx, query, args...)
	}

	if err != nil {
		return apperrors.Internal(fmt.Errorf("exec: %w", err))
	}

	return nil
}

// DisconnectBeerAndFeature удаляет связь характеристики с сортом пива. Если связи нет, возвращает ошибку.
func (r *BeerPostgres) DisconnectBeerAndFeature(ctx context.Context, tx pgx.Tx, beerID uint) error {
	if r.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.DisconnectBeerAndFeature(beerID)
	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	if tx != nil {
		_, err = tx.Exec(ctx, query, args...)
	} else {
		_, err = r.Pool.Exec(ctx, query, args...)
	}

	if err != nil {
		return apperrors.Internal(fmt.Errorf("exec: %w", err))
	}

	return nil
}

// GetCountryID возвращает ID страны по ее названиюс возвожностью запуска в транзакции. Если страны нет, она будет добавлена в базу данных. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) GetCountryID(ctx context.Context, tx pgx.Tx, name string) (uint, error) {
	if name == "" {
		return 0, apperrors.BadRequest("country name is empty", errors.New("country name is empty"))
	}
	if r.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectOrInsertCountry(name)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, args...)
	} else {
		row = r.Pool.QueryRow(ctx, query, args...)
	}

	var countryID uint
	if err := row.Scan(&countryID); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("scan: %w", err))
	}
	return countryID, nil
}

// GetCityID возвращает ID существующего или созданного города по его названию с возвожностью запуска в транзакции
func (r *BeerPostgres) GetCityID(ctx context.Context, tx pgx.Tx, name string, countryID uint) (uint, error) {
	if name == "" {
		return 0, apperrors.BadRequest("city name is empty", errors.New("city name is empty"))
	}
	if r.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectOrInsertCity(name, countryID)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, args...)
	} else {
		row = r.Pool.QueryRow(ctx, query, args...)
	}

	var cityID uint
	if err = row.Scan(&cityID); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("city QueryRow: %w", err))
	}

	return cityID, nil
}

// GetFeatureID возвращает ID характеристики по ее названию. Если характеристики нет, она будет добавлена в базу данных. Если name пустой, возвращает ошибку.
func (r *BeerPostgres) GetFeatureID(ctx context.Context, tx pgx.Tx, name string) (uint, error) {
	if r.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectOrInsertFeature(name)
	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("toSql: %w", err))
	}

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, args...)
	} else {
		row = r.Pool.QueryRow(ctx, query, args...)
	}

	var featID uint
	if err = row.Scan(&featID); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("city QueryRow: %w", err))
	}
	return featID, nil
}

// scanBeer сканирует строку из базы данных в сущность Beer. Если строка не соответствует структуре сущности, возвращает ошибку.
func scanBeer(row pgx.Row) (*entities.Beer, error) {
	var beer entities.Beer
	var reviewRatingSum, reviewAmount uint
	err := row.Scan(&beer.ID, &beer.Name,
		&beer.Description, &beer.ABV, &beer.IBU, &beer.Amount,
		&beer.Unit, &beer.City, &beer.Country,
		&beer.Category.Name, &beer.Features,
		&reviewRatingSum, &reviewAmount)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("%s: %w", "Scan", err))
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
		return nil, apperrors.Internal(fmt.Errorf("%s: %w", "Scan", err))
	}

	return &beer, nil
}
