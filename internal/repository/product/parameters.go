package product

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func (repo *Repository) productFromRow(row queries.GetProductByIDRow) entities.Product {
	product := repo.mapper.FromDBFullProduct(row)
	if row.Units != nil {
		product.Unit = *row.Units
	}
	return product
}

func (repo *Repository) getProduct(ctx context.Context, q *queries.Queries, id int64) (*entities.Product, error) {
	row, err := q.GetProductByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.NotFound("product not found", err)
	}
	if err != nil {
		return nil, err
	}
	products := []entities.Product{repo.productFromRow(row)}
	if err := repo.loadParameters(ctx, q, products); err != nil {
		return nil, err
	}
	return &products[0], nil
}

// loadParameters загружает значения всех изделий одним запросом.
func (repo *Repository) loadParameters(ctx context.Context, q *queries.Queries, products []entities.Product) error {
	if len(products) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(products))
	byID := make(map[int64]map[string]any, len(products))
	for i := range products {
		products[i].Parameters = make(map[string]any)
		ids = append(ids, products[i].ID)
		byID[products[i].ID] = products[i].Parameters
	}
	rows, err := q.GetProductParameterValues(ctx, ids)
	if err != nil {
		return fmt.Errorf("get product parameters: %w", err)
	}
	for _, row := range rows {
		value, err := (&entities.ParameterValueRow{ValueRaw: row.ValueRaw, ValueType: entities.EnumType(row.ValueType)}).FromRow()
		if err != nil {
			return fmt.Errorf("decode parameter %q of product %d: %w", row.FieldName, row.ProductID, err)
		}
		if _, exists := byID[row.ProductID][row.FieldName]; exists {
			return fmt.Errorf("ambiguous parameter field %q for product %d", row.FieldName, row.ProductID)
		}
		byID[row.ProductID][row.FieldName] = value.Value
	}
	return nil
}

func saveParameters(ctx context.Context, q *queries.Queries, productID, classID int64, values []entities.ProductParameterValue) error {
	if len(values) == 0 {
		return nil
	}
	class, err := q.GetClassByID(ctx, classID)
	if err != nil {
		return err
	}
	for _, value := range values {
		if value.RangeParameterID != nil {
			if value.NumericValue == nil || value.EnumParameterID != nil || value.EnumValueID != nil {
				return errors.New("invalid numeric parameter assignment")
			}
			if !containsID(class.NumericParameterIds, *value.RangeParameterID) {
				return apperrors.Conflict("class parameters changed; reload the product and retry", nil)
			}
			err = q.InsertProductRangeValue(ctx, queries.InsertProductRangeValueParams{
				ProductID: productID, RangeParameterID: value.RangeParameterID,
				NumericValueRaw: strconv.FormatFloat(*value.NumericValue, 'f', 3, 64),
			})
		} else {
			if value.EnumParameterID == nil || value.EnumValueID == nil || value.NumericValue != nil {
				return errors.New("invalid enum parameter assignment")
			}
			if !containsID(class.EnumParameterIds, *value.EnumParameterID) {
				return apperrors.Conflict("class parameters changed; reload the product and retry", nil)
			}
			err = q.InsertProductEnumValue(ctx, queries.InsertProductEnumValueParams{
				ProductID: productID, EnumParameterID: value.EnumParameterID, EnumValueID: value.EnumValueID,
			})
		}
		if err != nil {
			return fmt.Errorf("save product parameter: %w", err)
		}
	}
	return nil
}

func containsID(ids []int64, id int64) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}
