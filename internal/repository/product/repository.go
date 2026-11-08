package product

import (
	"Brewery/internal/entities"
	mpr "Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	Pool   *pgxpool.Pool
	query  *queries.Queries
	mapper mpr.ProductMapperImpl
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
	product = repo.mapper.FromInsert(productRow)
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
	products := repo.mapper.FromDBProducts(dbProducts)
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
	products := repo.mapper.FromDBFullProduct(dbProduct)
	return &products, nil
}

func (repo *Repository) UpdateProduct(ctx context.Context, id int64, updates map[string]any) (*entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	tx, err := repo.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qTx := repo.query.WithTx(tx)

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
	var baseID int64 = -1
	if val, ok := updates["base_id"]; ok {
		if baseID, ok := val.(int64); ok {
			params.BaseID = &baseID
		}
	}

	dbProduct, err := qTx.UpdateProduct(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}

	// При обновлении base_id необходимо поменять все необходимые материалы со старого изделия на новое
	if baseID != -1 {
		if err := qTx.DeleteProductsMaterials(ctx, id); err != nil {
			return nil, fmt.Errorf("DeleteProductsMaterials: %w", err)
		}
		params := queries.CopyBaseMaterialsParams{
			ProductID: id,
			BaseID:    baseID,
		}
		err := qTx.CopyBaseMaterials(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("CopyBaseMaterials: %w", err)
		}
	}
	products := repo.mapper.FromUpdate(dbProduct)
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
	products := repo.mapper.FromDBProducts(dbProduct)
	return products, nil
}

func (repo *Repository) GetFilteredProducts(
	ctx context.Context, filters []entities.FilterParameter,
	classID *uint, limit, offset int32,
) ([]entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	builder := squirrel.Select(
		"id",
		"name",
		"amount",
		"units",
		"class_id",
		"base_id",
	).From("products").
		PlaceholderFormat(squirrel.Dollar)

	if len(filters) != 0 {
		for _, filter := range filters {
			field := filter.FieldName
			val := filter.Value
			var pred any

			oper := filter.Operation
			switch oper {
			case entities.OpEqual:
				pred = squirrel.Eq{field: val}
			case entities.OpGreater:
				pred = squirrel.Gt{field: val}
			case entities.OpGreaterEqual:
				pred = squirrel.GtOrEq{field: val}
			case entities.OpLess:
				pred = squirrel.Lt{field: val}
			case entities.OpLessEqual:
				pred = squirrel.LtOrEq{field: val}
			case entities.OpNotEqual:
				pred = squirrel.NotEq{field: val}
			}

			builder = builder.Where(pred)
		}
	}

	if classID != nil {
		builder = builder.Where(squirrel.Eq{"product_id": *classID})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	rows, err := repo.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	var filteredProducts []queries.Product
	for rows.Next() {
		var p queries.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Amount, &p.Units, &p.ClassID, &p.BaseID); err != nil {
			return nil, err
		}
		filteredProducts = append(filteredProducts, p)
	}
	return repo.mapper.FromDBProducts(filteredProducts), nil
}
