package product

import (
	"Brewery/internal/entities"
	mpr "Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var mapper = mpr.ProductMapperImpl{}

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

func (repo *Repository) InsertProduct(ctx context.Context, product entities.Product) (*entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	params := queries.InsertProductParams{
		Name:   product.Name,
		Amount: &product.Amount,
		Units:  &product.Unit,
		BaseID: &product.BaseID,
	}
	productRow, err := repo.query.InsertProduct(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	product = mapper.FromInsert(productRow)
	return &product, nil
}

func (repo *Repository) GetProducts(ctx context.Context, limit, offset int32) ([]entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	params := queries.GetProductsParams{
		Limit:  limit,
		Offset: offset,
	}
	dbProducts, err := repo.query.GetProducts(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	products := mapper.FromDBProducts(dbProducts)
	return products, nil
}

func (repo *Repository) GetProductByID(ctx context.Context, id int64) (*entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	dbProduct, err := repo.query.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	products := mapper.FromDBFullProduct(dbProduct)
	return &products, nil
}

func (repo *Repository) UpdateProduct(ctx context.Context, id int64, updates map[string]any) (*entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	params := queries.UpdateProductParams{ID: id}
	if val, ok := updates["name"]; ok {
		if name, ok := val.(string); ok {
			params.Name = &name
		}
	}
	if val, ok := updates["amount"]; ok {
		if amount, ok := val.(float32); ok {
			params.Amount = &amount
		}
	}
	if val, ok := updates["units"]; ok {
		if units, ok := val.(float32); ok {
			params.Amount = &units
		}
	}
	if val, ok := updates["class_id"]; ok {
		if classID, ok := val.(int64); ok {
			params.ClassID = &classID
		}
	}
	if val, ok := updates["base_id"]; ok {
		if baseID, ok := val.(int64); ok {
			params.BaseID = &baseID
		}
	}

	dbProduct, err := repo.query.UpdateProduct(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	products := mapper.FromUpdate(dbProduct)
	return &products, nil
}

func (repo *Repository) DeleteProduct(ctx context.Context, id int64) error {
	if repo.Pool == nil {
		return errors.New("db pool is empty")
	}
	err := repo.query.DeleteProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("InsertProduct: %w", err)
	}
	return nil
}

func (repo *Repository) GetProductsByClass(
	ctx context.Context, classID int64, limit, offset int32,
) ([]entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	dbProduct, err := repo.query.GetProductByClassID(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	products := mapper.FromDBProducts(dbProduct)
	return products, nil
}

func (repo *Repository) GetFilteredProducts(
	ctx context.Context, filters []entities.FilterParameter,
	classID *uint, limit, offset int32,
) ([]entities.Product, error)
