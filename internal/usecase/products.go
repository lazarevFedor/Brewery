package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/product"
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context, product entities.Product) (*entities.Product, error)
	GetProducts(ctx context.Context, limit, offset int32) ([]entities.Product, error)
	GetProductByID(ctx context.Context, id int64) (*entities.Product, error)
	UpdateProduct(ctx context.Context, id int64, updates map[string]any) (*entities.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProductsByClass(ctx context.Context, classID int64, limit, offset int32) ([]entities.Product, error)
	GetFilteredProducts(ctx context.Context, filters []entities.FilterParameter, classID *uint, limit int32, offset int32) ([]entities.Product, error)
}

// NewProductRepository создает новый экземпляр ProductRepository с использованием предоставленного пула соединений к базе данных.
func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &product.Repository{Pool: pool}
}

type ProductService struct {
	repo         ProductRepository
	enumRepo     EnumRepository
	paramRepo    ParameterRepository
	classRepo    ClassRepository
	materialRepo MaterialRepository
}

// NewProductService creates product service with all dependencies.
func NewProductService(
	repo ProductRepository,
	enumRepo EnumRepository,
	paramRepo ParameterRepository,
	classRepo ClassRepository,
	materialRepo MaterialRepository,
) *ProductService {
	return &ProductService{
		repo:         repo,
		enumRepo:     enumRepo,
		paramRepo:    paramRepo,
		classRepo:    classRepo,
		materialRepo: materialRepo,
	}
}

func (ps *ProductService) ensureRepo() error {
	if ps == nil {
		return errors.New("product service is nil")
	}
	if ps.repo == nil {
		return errors.New("product repository is nil")
	}
	if ps.classRepo == nil {
		return errors.New("class repository is nil")
	}
	if ps.paramRepo == nil {
		return errors.New("parameter repository is nil")
	}
	if ps.enumRepo == nil {
		return errors.New("enum repository is nil")
	}
	return nil
}

func (ps *ProductService) CreateProduct(ctx context.Context, p entities.Product) (*entities.Product, error) {
	if err := ps.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	p, err := ps.ValidateProductCreation(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to validate product creation: %w", err)
	}

	err = ps.ValidateProductWithParameters(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to validate product: %w", err)
	}

	createdProduct, err := ps.repo.InsertProduct(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	if createdProduct == nil {
		return nil, errors.New("failed to create product: returned product is nil")
	}

	return createdProduct, nil
}

func (ps *ProductService) GetProducts(ctx context.Context, limit, offset int32) ([]entities.Product, error) {
	if err := ps.ensureRepo(); err != nil {
		return nil, err
	}
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
	if err := ps.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, errors.New("invalid product id")
	}

	_, err := ps.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	err = ps.ValidateUpdates(ctx, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to validate updates: %w", err)
	}

	p, err := ps.repo.UpdateProduct(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return p, nil
}

func (ps *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	if err := ps.ensureRepo(); err != nil {
		return err
	}
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
	if err := ps.ensureRepo(); err != nil {
		return nil, err
	}
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
	if err := ps.ensureRepo(); err != nil {
		return nil, err
	}
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
	if err := ps.ensureRepo(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, errors.New("invalid product id")
	}

	p, err := ps.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return p, nil
}

func (ps *ProductService) ValidateUpdates(ctx context.Context, updates map[string]any) error {
	var baseIDInt int64

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	for key := range updates {
		switch key {
		case "name":
			name, _ := updates[key]
			if nameStr, ok := name.(string); !ok || nameStr == "" {
				return errors.New("invalid product name")
			}
		case "class_id":
			classID, _ := updates[key]
			switch v := classID.(type) {
			case int, int8, int16, int32, int64:
				if toInt64(v) == 0 {
					return errors.New("invalid class id")
				}
			default:
				return errors.New("invalid class id")
			}
		case "base_id":
			baseID, _ := updates[key]
			if baseIDInt, ok := toInt64Value(baseID); !ok || baseIDInt == 0 {
				return errors.New("invalid base id")
			}
			_, err := ps.repo.GetProductByID(ctx, baseIDInt)
			if err != nil {
				return fmt.Errorf("base product with id %d does not exist: %w", baseIDInt, err)
			}
		case "amount":
			amount, _ := updates[key]
			if _, ok := toFloat64Value(amount); !ok {
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

func toInt64Value(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int8:
		return int64(x), true
	case int16:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	case uint:
		return int64(x), true
	case uint8:
		return int64(x), true
	case uint16:
		return int64(x), true
	case uint32:
		return int64(x), true
	case uint64:
		return int64(x), true
	case float32:
		if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
			return 0, false
		}
		return int64(x), true
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return 0, false
		}
		return int64(x), true
	default:
		return 0, false
	}
}

func toFloat64Value(v any) (float64, bool) {
	switch x := v.(type) {
	case float32:
		return float64(x), true
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int8:
		return float64(x), true
	case int16:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint8:
		return float64(x), true
	case uint16:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	default:
		return 0, false
	}
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint64:
		return int64(x)
	case float32:
		return int64(x)
	case float64:
		return int64(x)
	default:
		return 0
	}
}

func normalizeEnumValue(value any) any {
	switch v := value.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	default:
		return value
	}
}

func (ps *ProductService) ValidateProductWithParameters(ctx context.Context, product entities.Product) error {
	for _, id := range product.ProdClass.EnumParamIDs {
		enumValues, err := ps.enumRepo.GetEnumValuesByClassID(ctx, uint(id))
		if err != nil {
			return fmt.Errorf("failed to get enum values for class ID %d: %w", id, err)
		}

		if len(enumValues) == 0 {
			continue
		}

		enumClass, err := ps.enumRepo.GetEnumClassByID(ctx, uint(id))
		if err != nil {
			return fmt.Errorf("failed to get enum class for ID %d: %w", id, err)
		}
		if enumClass == nil {
			return fmt.Errorf("enum class for ID %d not found", id)
		}

		parameterValue, exists := product.Parameters[enumClass.FieldName]
		if !exists {
			continue
		}

		normalizedValue := normalizeEnumValue(parameterValue)
		if !containsEnumValue(enumValues, normalizedValue) {
			return fmt.Errorf(
				"invalid value %v for parameter %q",
				normalizedValue,
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
			continue
		}

		numberValue, ok := toFloat64Value(parameterValue)
		if !ok {
			return fmt.Errorf("parameter %q must be a number", rangeParam.FieldName)
		}

		if numberValue < float64(rangeParam.MinValue) || numberValue > float64(rangeParam.MaxValue) {
			return fmt.Errorf(
				"parameter %q value %v is out of range [%v, %v]",
				rangeParam.FieldName,
				numberValue,
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

func (ps *ProductService) ValidateProductCreation(ctx context.Context, p entities.Product) (entities.Product, error) {
	if p.Name == "" {
		return p, errors.New("product name is required")
	}

	class, err := ps.classRepo.GetClassByName(ctx, p.ProdClass.Name)
	if err != nil {
		return p, fmt.Errorf("failed to get class by name: %w", err)
	}

	if class == nil {
		return p, fmt.Errorf("class with name %s does not exist", p.ProdClass.Name)
	}

	p.ProdClass = *class

	if p.BaseID != 0 {
		_, err = ps.repo.GetProductByID(ctx, p.BaseID)
		if err != nil {
			return p, fmt.Errorf("base product with id %d does not exist: %w", p.BaseID, err)
		}
	}

	return p, nil
}
