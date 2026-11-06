package usecase

import (
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"
	"reflect"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context, product entities.Product) (*entities.Product, error)
	GetProducts(ctx context.Context, limit, offset uint64) ([]entities.Product, error)
	GetProductByID(ctx context.Context, id int64) (*entities.Product, error)
	UpdateProduct(ctx context.Context, id int64, updates map[string]any) (*entities.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProductsByClass(ctx context.Context, classID int64, limit, offset int32) ([]entities.Product, error)
	GetFilteredProducts(ctx context.Context, filters []entities.FilterParameter, classID *uint, limit int32, offset int32) ([]entities.Product, error)
}

type ProductService struct {
	repo      ProductRepository
	enumRepo  EnumRepository
	paramRepo ParameterRepository
}

func (ps *ProductService) CreateProduct(ctx context.Context, product entities.Product) (*entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if product.Name == "" {
		return nil, errors.New("product name is required")
	}

	err := ps.ValidateProductWithParameters(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to validate product: %w", err)
	}

	createdProduct, err := ps.repo.InsertProduct(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return createdProduct, nil
}

func (ps *ProductService) GetProducts(ctx context.Context, limit, offset uint64) ([]entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	products, err := ps.repo.GetProducts(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return products, nil
}

func (ps *ProductService) UpdateProduct(ctx context.Context, id int64, updates map[string]any) (*entities.Product, error) {
	err := ps.ValidateUpdates(updates)
	if err != nil {
		return nil, fmt.Errorf("failed to validate updates: %w", err)
	}

	product, err := ps.repo.UpdateProduct(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (ps *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid product id")
	}

	err := ps.repo.DeleteProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (ps *ProductService) GetProductsByClass(ctx context.Context, classID int64, limit, offset int32) ([]entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if classID == 0 {
		return nil, errors.New("invalid class id")
	}

	products, err := ps.repo.GetProductsByClass(ctx, classID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by category: %w", err)
	}

	return products, nil
}

func (ps *ProductService) GetFilteredProducts(ctx context.Context, filters []entities.FilterParameter, classID *uint, limit, offset int32) ([]entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	products, err := ps.repo.GetFilteredProducts(ctx, filters, classID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return products, nil
}

func (ps *ProductService) GetProductByID(ctx context.Context, id int64) (*entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	product, err := ps.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (ps *ProductService) ValidateUpdates(updates map[string]any) error {
	for key := range updates {
		switch key {
		case "name":
			name, _ := updates[key]
			if nameStr, ok := name.(string); !ok || nameStr == "" {
				return errors.New("invalid product name")
			}
		case "class_id":
			classID, _ := updates[key]
			if classIDUint, ok := classID.(int64); !ok || classIDUint == 0 {
				return errors.New("invalid class id")
			}
		case "base_id":
			baseID, _ := updates[key]
			if baseIDUint, ok := baseID.(int); !ok || baseIDUint == 0 {
				return errors.New("invalid base id")
			}
		case "amount":
			amount, _ := updates[key]
			if amountUint, ok := amount.(uint); !ok || amountUint == 0 {
				return errors.New("invalid amount")
			}
		case "units":
			units, _ := updates[key]
			if unitsStr, ok := units.(string); !ok || unitsStr == "" {
				return errors.New("invalid units")
			}
		default:
			return fmt.Errorf("invalid update field: %s", key)
		}
	}

	return nil
}

func (ps *ProductService) ValidateProductWithParameters(ctx context.Context, product entities.Product) error {
	for id := range product.ProdClass.EnumParamIDs {
		enumValues, err := ps.enumRepo.GetEnumValuesByClassID(ctx, uint(id))
		if err != nil {
			return fmt.Errorf("failed to get enum values for class ID %d: %w", id, err)
		}
		enumClass, err := ps.enumRepo.GetEnumClassByID(ctx, uint(id))
		if err != nil {
			return fmt.Errorf("failed to get enum class for ID %d: %w", id, err)
		}

		parameterValue, exists := product.Parameters[enumClass.FieldName]
		if !exists {
			return fmt.Errorf("parameter %q is required", enumClass.FieldName)
		}

		if !containsEnumValue(enumValues, parameterValue) {
			return fmt.Errorf(
				"invalid value %v for parameter %q",
				parameterValue,
				enumClass.FieldName,
			)
		}

	}

	rangeParameters, _, err := ps.paramRepo.GetParameters(ctx, product.ProdClass.ID, entities.RangeParameterType)
	if err != nil {
		return fmt.Errorf("failed to get range parameters for class ID %d: %w", product.ProdClass.ID, err)
	}
	for _, rangeParam := range rangeParameters {
		parameterValue, exists := product.Parameters[rangeParam.FieldName]
		if !exists {
			return fmt.Errorf("parameter %q is required", rangeParam.FieldName)
		}

		parameterValueInt, ok := parameterValue.(int)
		if !ok {
			return fmt.Errorf("parameter %q must be a number", rangeParam.FieldName)
		}

		if parameterValueInt < rangeParam.MinValue || parameterValueInt > rangeParam.MaxValue {
			return fmt.Errorf(
				"parameter %q value %v is out of range [%v, %v]",
				rangeParam.FieldName,
				parameterValueInt,
				rangeParam.MinValue,
				rangeParam.MaxValue,
			)
		}
	}

	return nil
}

func containsEnumValue(enumValues []entities.EnumValue, parameterValue any) bool {
	for _, enumValue := range enumValues {
		if reflect.DeepEqual(enumValue.Value, parameterValue) {
			return true
		}
	}

	return false
}
