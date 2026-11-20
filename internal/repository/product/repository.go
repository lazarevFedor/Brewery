package product

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	mpr "Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

func (repo *Repository) InsertProduct(ctx context.Context, product entities.Product, values []entities.ParameterValue) (*entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	tx, err := repo.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qTx := repo.query.WithTx(tx)

	if err := repository.ValidateParameterAssignments(ctx, qTx, product.ProdClass.ID, values); err != nil {
		return nil, err
	}

	params := queries.InsertProductParams{
		Name:           product.Name,
		Amount:         &product.Amount,
		Units:          &product.Unit,
		ClassID:        product.ProdClass.ID,
		Price:          &product.Price,
		IsPurchased:    product.IsPurchased,
		IsFinalProduct: product.IsFinalProduct,
	}
	if product.BaseID != 0 {
		params.BaseID = &product.BaseID
	}
	productRow, err := qTx.InsertProduct(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("InsertProduct: %w", productWriteError(err))
	}
	if product.BaseID != 0 {
		if err := qTx.CopyBaseMaterials(ctx, queries.CopyBaseMaterialsParams{
			ProductID: productRow.ID,
			BaseID:    product.BaseID,
		}); err != nil {
			return nil, fmt.Errorf("CopyBaseMaterials: %w", err)
		}
	}
	if err := saveParameters(ctx, qTx, productRow.ID, values); err != nil {
		return nil, err
	}
	created, err := repo.getProduct(ctx, qTx, productRow.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit product creation: %w", err)
	}
	return created, nil
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
		products = append(products, repo.productFromRow(queries.GetProductByIDRow(row)))
	}
	if err := repo.loadParameters(ctx, repo.query, products); err != nil {
		return nil, err
	}
	return products, nil
}

func (repo *Repository) GetProductByID(ctx context.Context, id int64) (*entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	return repo.getProduct(ctx, repo.query, id)
}

func (repo *Repository) UpdateProduct(ctx context.Context, id int64, changes entities.ProductUpdate) (*entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	tx, err := repo.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qTx := repo.query.WithTx(tx)

	locked, err := qTx.LockProductForUpdate(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.NotFound("product not found", err)
	}
	if err != nil {
		return nil, err
	}
	classID := locked.ClassID
	if classID != changes.ExpectedClassID {
		return nil, apperrors.Conflict("product class changed; reload the product and retry", nil)
	}
	if changes.ReplaceParameters {
		current, err := repo.getProduct(ctx, qTx, id)
		if err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(current.Parameters, changes.ExpectedParameters) {
			return nil, apperrors.Conflict("product parameters changed; reload the product and retry", nil)
		}
	}
	updates := changes.Fields
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
	if newClassID, ok := updates["class_id"].(int64); ok {
		classID = newClassID
		params.ClassID = &newClassID
	}
	if value, present := updates["base_id"]; present {
		baseID, _ := value.(int64)
		currentBaseID := int64(0)
		if locked.BaseID != nil {
			currentBaseID = *locked.BaseID
		}
		if baseID != currentBaseID {
			if currentBaseID != changes.ExpectedBaseID {
				return nil, apperrors.Conflict("product base changed; reload the product and retry", nil)
			}
			params.SetBaseID = true
			if baseID != 0 {
				params.BaseID = &baseID
			}
		}
	}
	if price, ok := updates["price"].(float64); ok {
		params.Price = &price
	}
	if isPurchased, ok := updates["is_purchased"].(bool); ok {
		params.IsPurchased = pgtype.Bool{Bool: isPurchased, Valid: true}
	}
	if isFinalProduct, ok := updates["is_final_product"].(bool); ok {
		params.IsFinalProduct = pgtype.Bool{Bool: isFinalProduct, Valid: true}
	}

	if err := repository.ValidateParameterAssignments(ctx, qTx, classID, changes.ParameterValues); err != nil {
		return nil, err
	}
	if _, err := qTx.UpdateProduct(ctx, params); err != nil {
		return nil, fmt.Errorf("UpdateProduct: %w", productWriteError(err))
	}

	if params.SetBaseID && params.BaseID != nil {
		if err := qTx.SyncBaseMaterials(ctx, queries.SyncBaseMaterialsParams{ProductID: id, BaseID: *params.BaseID}); err != nil {
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.ConstraintName == "fk_operations_material" {
				return nil, apperrors.Conflict("product composition is referenced by an operation", err)
			}
			return nil, fmt.Errorf("SyncBaseMaterials: %w", err)
		}
	}
	if changes.ReplaceParameters {
		err = qTx.DeleteAllProductParameters(ctx, id)
	} else if len(changes.ParameterFields) > 0 {
		err = qTx.DeleteProductParameterFields(ctx, queries.DeleteProductParameterFieldsParams{ProductID: id, FieldNames: changes.ParameterFields})
	}
	if err != nil {
		return nil, err
	}
	if err := saveParameters(ctx, qTx, id, changes.ParameterValues); err != nil {
		return nil, err
	}
	updated, err := repo.getProduct(ctx, qTx, id)
	if err != nil {
		return nil, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("Commit: %w", err)
	}

	return updated, nil
}

func productWriteError(err error) error {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok {
		return err
	}

	switch pgErr.ConstraintName {
	case "products_class_id_fkey":
		return apperrors.BadRequest("product class does not exist", err)
	case "products_base_id_fkey":
		return apperrors.BadRequest("base product does not exist", err)
	case "chk_product_isnot_its_base":
		return apperrors.BadRequest("product cannot be its own base", err)
	case "check_product_amount_positive":
		return apperrors.BadRequest("invalid amount", err)
	case "chk_product_price_positive":
		return apperrors.BadRequest("invalid price", err)
	}
	if pgErr.Code == "22003" {
		return apperrors.BadRequest("product number is out of range", err)
	}
	return err
}

func (repo *Repository) DeleteProduct(ctx context.Context, id int64) error {
	if repo.Pool == nil {
		return errors.New("db pool is empty")
	}
	count, err := repo.query.DeleteProduct(ctx, id)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.ConstraintName {
			case "operations_input_product_id_fkey", "operations_output_product_id_fkey", "fk_operations_material":
				return apperrors.Conflict("product is referenced by an operation", err)
			}
		}
		return fmt.Errorf("DeleteProduct: %w", err)
	}
	if count == 0 {
		return apperrors.NotFound("product not found", nil)
	}
	return nil
}

func (repo *Repository) GetProductsByClass(
	ctx context.Context, classID int64, limit, offset int32,
) ([]entities.Product, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	dbProduct, err := repo.query.GetProductByClassID(ctx, queries.GetProductByClassIDParams{ClassID: classID, LimitCount: limit, OffsetCount: offset})
	if err != nil {
		return nil, fmt.Errorf("GetProductsByClass: %w", err)
	}
	products := make([]entities.Product, 0, len(dbProduct))
	for _, row := range dbProduct {
		products = append(products, repo.productFromRow(queries.GetProductByIDRow(row)))
	}
	if err := repo.loadParameters(ctx, repo.query, products); err != nil {
		return nil, err
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
		"COALESCE(p.price, 0)::double precision AS price",
		"p.is_purchased",
		"p.is_final_product",
		"c.id AS class_id",
		"c.name AS class_name",
		"c.parent_id AS class_parent_id",
		"c.numeric_parameter_ids AS class_num_params",
		"c.enum_parameter_ids AS class_enum_params",
	).From("products p").
		Join("classes c ON p.class_id = c.id").
		PlaceholderFormat(squirrel.Dollar)

	for _, filter := range filters {
		predicate, err := productFilter(filter)
		if err != nil {
			return nil, err
		}
		builder = builder.Where(predicate)
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
			&p.ID, &p.Name, &p.Amount, &p.Units, &p.BaseID, &p.Price, &p.IsPurchased, &p.IsFinalProduct,
			&p.ClassID, &p.ClassName, &p.ClassParentID, &p.ClassNumParams, &p.ClassEnumParams,
		); err != nil {
			return nil, err
		}
		filteredProducts = append(filteredProducts, repo.productFromRow(p))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read filtered products: %w", err)
	}
	rows.Close()
	if err := repo.loadParameters(ctx, repo.query, filteredProducts); err != nil {
		return nil, err
	}
	return filteredProducts, nil
}
