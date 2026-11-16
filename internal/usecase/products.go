package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/product"
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context, product entities.Product, values []entities.ProductParameterValue) (*entities.Product, error)
	GetProducts(ctx context.Context, limit, offset int32) ([]entities.Product, error)
	GetProductByID(ctx context.Context, id int64) (*entities.Product, error)
	UpdateProduct(ctx context.Context, id int64, updates entities.ProductUpdate) (*entities.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	GetProductsByClass(ctx context.Context, classID int64, limit, offset int32) ([]entities.Product, error)
	GetFilteredProducts(ctx context.Context, filters []entities.FilterParameter, classID *uint, limit int32, offset int32) ([]entities.Product, error)
}

// NewProductRepository создает новый экземпляр ProductRepository с использованием предоставленного пула соединений к базе данных.
func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return product.NewRepository(pool)
}

type ProductService struct {
	repo                ProductRepository
	enumValuesRepo      EnumValuesRepository
	classParametersRepo ClassParametersRepository
	classRepo           ClassRepository
	materialRepo        MaterialRepository
}

// NewProductService creates product service with all dependencies.
func NewProductService(
	repo ProductRepository,
	enumValuesRepo EnumValuesRepository,
	classParametersRepo ClassParametersRepository,
	classRepo ClassRepository,
	materialRepo MaterialRepository,
) *ProductService {
	return &ProductService{
		repo:                repo,
		enumValuesRepo:      enumValuesRepo,
		classParametersRepo: classParametersRepo,
		classRepo:           classRepo,
		materialRepo:        materialRepo,
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
	if ps.classParametersRepo == nil {
		return errors.New("class parameters repository is nil")
	}
	if ps.enumValuesRepo == nil {
		return errors.New("enum values repository is nil")
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

	values, canonical, err := ps.resolveProductParameters(ctx, p)
	if err != nil {
		return nil, err
	}
	p.Parameters = canonical
	createdProduct, err := ps.repo.InsertProduct(ctx, p, values)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	if createdProduct == nil {
		return nil, errors.New("failed to create product: returned product is nil")
	}

	createdProduct.ProdClass = p.ProdClass
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

	current, err := ps.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	err = ps.ValidateUpdates(ctx, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to validate updates: %w", err)
	}

	changes := entities.ProductUpdate{Fields: maps.Clone(updates), ExpectedClassID: current.ProdClass.ID}
	delete(changes.Fields, "parameters")
	candidate := *current
	candidate.Parameters = maps.Clone(current.Parameters)
	if candidate.Parameters == nil {
		candidate.Parameters = make(map[string]any)
	}
	if classID, ok := updates["class_id"].(int64); ok && classID != current.ProdClass.ID {
		class, err := ps.classRepo.GetClassByID(ctx, classID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.BadRequest("class does not exist", err)
		}
		if err != nil {
			return nil, fmt.Errorf("get new product class: %w", err)
		}
		if class == nil {
			return nil, apperrors.BadRequest("class does not exist", nil)
		}
		candidate.ProdClass = *class
		changes.ReplaceParameters = true
		changes.ExpectedParameters = maps.Clone(current.Parameters)
	}
	patch, hasParameters := updates["parameters"]
	if hasParameters {
		for field, value := range patch.(map[string]any) {
			changes.ParameterFields = append(changes.ParameterFields, field)
			if value == nil {
				delete(candidate.Parameters, field)
			} else {
				candidate.Parameters[field] = value
			}
		}
		slices.Sort(changes.ParameterFields)
	}
	if hasParameters || changes.ReplaceParameters {
		values, _, err := ps.resolveProductParameters(ctx, candidate)
		if err != nil {
			return nil, err
		}
		if changes.ReplaceParameters {
			changes.ParameterValues = values
		} else {
			for _, value := range values {
				if slices.Contains(changes.ParameterFields, value.FieldName) {
					changes.ParameterValues = append(changes.ParameterValues, value)
				}
			}
		}
	}

	if baseID, ok := updates["base_id"].(int64); ok {
		if baseID == id {
			return nil, errors.New("product cannot be its own base")
		}
		if ps.materialRepo == nil {
			return nil, errors.New("material repository is nil")
		}
		materials, err := ps.materialRepo.GetMaterialsByProduct(ctx, baseID)
		if err != nil {
			return nil, fmt.Errorf("failed to get base product materials: %w", err)
		}
		materialIDs := make([]int64, 0, len(materials))
		for _, material := range materials {
			materialID := material.MaterialID
			if materialID == 0 {
				materialID = material.ID
			}
			materialIDs = append(materialIDs, materialID)
		}
		if err := validateMaterialComposition(ctx, ps.materialRepo, id, materialIDs); err != nil {
			return nil, fmt.Errorf("failed to validate base product composition: %w", err)
		}
	}

	p, err := ps.repo.UpdateProduct(ctx, id, changes)
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
				if toInt64(v) <= 0 {
					return apperrors.BadRequest("invalid class id", nil)
				}
				updates[key] = toInt64(v)
			default:
				return errors.New("invalid class id")
			}
		case "base_id":
			baseID, _ := updates[key]
			baseIDInt, ok := toInt64Value(baseID)
			if !ok || baseIDInt <= 0 {
				return errors.New("invalid base id")
			}
			_, err := ps.repo.GetProductByID(ctx, baseIDInt)
			if err != nil {
				return fmt.Errorf("base product with id %d does not exist: %w", baseIDInt, err)
			}
			updates[key] = baseIDInt
		case "amount":
			amount, _ := updates[key]
			amountValue, ok := toFloat64Value(amount)
			if !ok {
				return errors.New("invalid amount")
			}
			updates[key] = float32(amountValue)
		case "parameters":
			if value, ok := updates[key].(map[string]any); !ok || value == nil {
				return apperrors.BadRequest("parameters must be an object; use null for an individual value to remove it", nil)
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

// ValidateProductWithParameters проверяет значения параметров по классу изделия.
func (ps *ProductService) ValidateProductWithParameters(ctx context.Context, product entities.Product) error {
	_, _, err := ps.resolveProductParameters(ctx, product)
	return err
}

func (ps *ProductService) ValidateProductCreation(ctx context.Context, p entities.Product) (entities.Product, error) {
	if p.Name == "" {
		return p, errors.New("product name is required")
	}

	var class *entities.Class
	var err error
	if p.ProdClass.ID != 0 {
		class, err = ps.classRepo.GetClassByID(ctx, p.ProdClass.ID)
	} else {
		class, err = ps.classRepo.GetClassByName(ctx, p.ProdClass.Name)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return p, apperrors.BadRequest("product class does not exist", err)
	}
	if err != nil {
		return p, fmt.Errorf("failed to get product class: %w", err)
	}
	if class == nil {
		return p, apperrors.BadRequest("product class does not exist", nil)
	}
	if p.ProdClass.Name != "" && p.ProdClass.Name != class.Name {
		return p, apperrors.BadRequest("class_id and class.name refer to different classes", nil)
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
