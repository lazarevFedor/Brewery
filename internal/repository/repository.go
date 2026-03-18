package postgres

import (
	"Brewery/internal/entities"
	"context"
	"fmt"

	// sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "embed"
)

var(
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

// Repository хранит в себе пул подлючений к БД
type Repository struct {
	pool *pgxpool.Pool
}


// NewRepository создает новый репозиторий БД
func NewRepository(pgPool *pgxpool.Pool) *Repository {
	return &Repository{pool: pgPool}
}


// InsertBeer - метод Repository для вставки объекта Beer в БД
func (r *Repository) InsertBeer(ctx context.Context, beer entities.Beer) error {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var countryID int
	err = tx.QueryRow(ctx, getOrCreateCountryQuery, beer.Country).Scan(&countryID)
	if err != nil {
		return fmt.Errorf("Country QueryRow: %w", err)
	}
	var cityID int
	err = tx.QueryRow(ctx, getOrCreateCityQuery, beer.City, countryID).Scan(&cityID)
	if err != nil {
		return fmt.Errorf("City QueryRow: %w", err)
	}
	var typeID int
	err = tx.QueryRow(ctx, getOrCreateTypeQuery, beer.Type).Scan(&typeID)
	if err != nil {
		return fmt.Errorf("Type QueryRow: %w", err)
	}
	var categoryID int
	err = tx.QueryRow(ctx, getProductCategoryByNameQuery, beer.Category).Scan(&categoryID)
	if err != nil {
		return fmt.Errorf("Category QueryRow: %w", err)
	}

	var beerID int
	err = tx.QueryRow(ctx, insertBeerQuery, 
		beer.Name, beer.Rating, beer.Description, beer.ABV, beer.IBU, 
		typeID, cityID, categoryID,
	).Scan(&beerID)
	if err != nil {
		return fmt.Errorf("Beer QueryRow: %w", err)
	}

	for _, featName := range beer.Features{
		var featID int
		err = tx.QueryRow(ctx, insertFeatureQuery, featName).Scan(&featID)
		if err != nil {
			return fmt.Errorf("Feature QueryRow: %w", err)
		}
		tx.Exec(ctx, insertBeerFeatureQuery, beerID, featID)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	return nil
}

// TODO: Добавить ошибки приложения

// GetBeers - метод Repository для получения всех позиций пива из БД
func (r *Repository) GetBeers(ctx context.Context) ([]entities.Beer, error){
	rows, err := r.pool.Query(ctx, getBeersQuery)
	if err != nil{
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close() 

	beers := make([]entities.Beer, 0)

	for rows.Next() {
		b := entities.Beer{}
		rows.Scan(&b.ID, &b.Name, &b.Rating, &b.Description,
			&b.ABV, &b.IBU, &b.City, &b.Country,
			&b.Category, &b.Type, &b.Features)

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
// 	row = r.pool.QueryRow(ctx, sql, args...)
	
// 	return nil
// }
