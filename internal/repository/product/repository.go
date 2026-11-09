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
	tx, err := repo.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qTx := repo.query.WithTx(tx)

	params := queries.InsertProductParams{
		Name:    product.Name,
		Amount:  &product.Amount,
		Units:   &product.Unit,
		ClassID: product.ProdClass.ID,
	}
	if product.BaseID != 0 {
		params.BaseID = &product.BaseID
	}
	productRow, err := qTx.InsertProduct(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("InsertProduct: %w", err)
	}
	if product.BaseID != 0 {
		if err := qTx.CopyBaseMaterials(ctx, queries.CopyBaseMaterialsParams{
			ProductID: productRow.ID,
			BaseID:    product.BaseID,
		}); err != nil {
			return nil, fmt.Errorf("CopyBaseMaterials: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit product creation: %w", err)
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
		return nil, fmt.Errorf("GetProducts: %w", err)
	}
	products := make([]entities.Product, 0, len(dbProducts))
	for _, row := range dbProducts {
		products = append(products, repo.mapper.FromDBFullProduct(queries.GetProductByIDRow(row)))
	}
	return products, nil
}

func (repo *Repository) GetProductByID(ctx context.Context, id int64) (*entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	dbProduct, err := repo.query.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetProductByID: %w", err)
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
		if units, ok := val.(string); ok {
			params.Units = &units
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

	dbProduct, err := qTx.UpdateProduct(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("UpdateProduct: %w", err)
	}

	// При обновлении base_id необходимо поменять все необходимые материалы со старого изделия на новое
	if params.BaseID != nil {
		if err := qTx.DeleteProductsMaterials(ctx, id); err != nil {
			return nil, fmt.Errorf("DeleteProductsMaterials: %w", err)
		}
		copyParams := queries.CopyBaseMaterialsParams{
			ProductID: id,
			BaseID:    *params.BaseID,
		}
		err := qTx.CopyBaseMaterials(ctx, copyParams)
		if err != nil {
			return nil, fmt.Errorf("CopyBaseMaterials: %w", err)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("Commit: %w", err)
	}

	products := repo.mapper.FromDBProduct(dbProduct)
	return &products, nil
}

func (repo *Repository) DeleteProduct(ctx context.Context, id int64) error {
	if repo.Pool == nil {
		return errors.New("db pool is empty")
	}
	err := repo.query.DeleteProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("DeleteProduct: %w", err)
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
		return nil, fmt.Errorf("GetProductsByClass: %w", err)
	}
	products := make([]entities.Product, 0, len(dbProduct))
	for _, row := range dbProduct {
		products = append(products, repo.mapper.FromDBFullProduct(queries.GetProductByIDRow(row)))
	}
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
		"p.id",
		"p.name",
		"p.amount",
		"p.units",
		"p.base_id",
		"c.id AS class_id",
		"c.name AS class_name",
		"c.parent_id AS class_parent_id",
		"c.numeric_parameter_ids AS class_num_params",
		"c.enum_parameter_ids AS class_enum_params",
	).From("products p").
		Join("classes c ON p.class_id = c.id").
		PlaceholderFormat(squirrel.Dollar)

	if len(filters) != 0 {
		for _, filter := range filters {
			field := "p." + filter.FieldName
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
		builder = builder.Where(squirrel.Eq{"p.class_id": *classID})
	}
	builder = builder.OrderBy("p.id").Limit(uint64(limit)).Offset(uint64(offset))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ToSql: %w", err)
	}

	rows, err := repo.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	var filteredProducts []entities.Product
	for rows.Next() {
		var p queries.GetProductByIDRow
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Amount, &p.Units, &p.BaseID,
			&p.ClassID, &p.ClassName, &p.ClassParentID, &p.ClassNumParams, &p.ClassEnumParams,
		); err != nil {
			return nil, err
		}
		filteredProducts = append(filteredProducts, repo.mapper.FromDBFullProduct(p))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read filtered products: %w", err)
	}
	return filteredProducts, nil
}
