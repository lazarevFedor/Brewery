package product

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	Pool  *pgxpool.Pool
	query *queries.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Pool:  pool,
		query: queries.New(pool),
	}
}

func (repo *Repository) InsertProduct(ctx context.Context, product entities.Product) (*entities.Product, error){
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	params := queries.InsertProductParams{
		Name: product.Name,
		Amount: &product.Amount,
		Units: &product.Unit,
		BaseID: &product.BaseID,
	}
	productRow, err := repo.query.InsertProduct(ctx, params)
	if err != nil{
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	product = mapper.Row2DomainProduct(productRow)
	return &product, nil
}

func (repo *Repository) GetProducts(ctx context.Context, limit, offset uint64) ([]entities.Product, error){
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	params := queries.GetProductsParams{
		Limit: int32(limit),
		Offset: int32(offset),
	}
	productRow, err := repo.query.GetProducts(ctx, params)
	if err != nil{
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	products := mapper.ToDomainProducts(productRow)
	return products, nil
}


func (repo *Repository) GetProductByID(ctx context.Context, id int64) (*entities.Product, error){
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	productRow, err := repo.query.GetProductByID(ctx, id)
	if err != nil{
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	products := mapper.ToDomainProducts(productRow)
	return products, nil
}

func (repo *Repository) UpdateProduct(ctx context.Context, updates map[string]any) (entities.Product, error)
func (repo *Repository) DeleteProduct(ctx context.Context, productID uint) error
func (repo *Repository) GetProductsByClass(
	ctx context.Context, classID uint, limit, offset uint64,
) (entities.Product, error)

func (repo *Repository) GetFilteredProducts(
	ctx context.Context, filters []entities.FilterParameter,
	classID *uint, limit, offset uint64,
) ([]entities.Product, error)
