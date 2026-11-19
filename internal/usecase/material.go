package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/material"
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MaterialRepository определяет интерфейс для работы с материалами в базе данных.
type MaterialRepository interface {
	// InsertMaterial вставляет новый материал в базу данных и возвращает его с присвоенным ID.
	InsertMaterial(ctx context.Context, material entities.Material) (*entities.Material, error)

	// UpdateMaterial обновляет существующий материал с заданным ID и возвращает обновленный материал.
	UpdateMaterial(ctx context.Context, material entities.Material) (*entities.Material, error)

	// DeleteMaterial удаляет материал с заданным ID из базы данных и возвращает удаленный материал.
	DeleteMaterial(ctx context.Context, productID, materialID int64) error

	// GetMaterialsByProduct извлекает все материалы, соответствующие заданному ID продукта, из базы данных.
	GetMaterialsByProduct(ctx context.Context, productID int64) ([]entities.MaterialProduct, error)

	// GetProductsByMaterial извлекает все продукты, использующие заданный материал, из базы данных.
	GetProductsByMaterial(ctx context.Context, materialID int64, limit, offset int32) ([]entities.MaterialProduct, error)
}

// NewMaterialRepository создает новый экземпляр MaterialRepository с использованием предоставленного пула соединений к базе данных.
func NewMaterialRepository(pool *pgxpool.Pool) MaterialRepository {
	return material.NewRepository(pool)
}

type MaterialService struct {
	repo        MaterialRepository
	productRepo ProductRepository
}

// NewMaterialService creates material service.
func NewMaterialService(repo MaterialRepository, productRepo ProductRepository) *MaterialService {
	return &MaterialService{repo: repo, productRepo: productRepo}
}

func (s *MaterialService) CreateMaterial(ctx context.Context, material entities.Material) (*entities.Material, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if material.ProductID <= 0 || material.MaterialID <= 0 {
		return nil, apperrors.BadRequest("invalid product id or material id", nil)
	}
	if material.ProductID == material.MaterialID {
		return nil, apperrors.BadRequest("product cannot be a material of itself", nil)
	}
	if !validMaterialAmount(material.MaterialAmount) {
		return nil, apperrors.BadRequest("material_amount must be a positive finite number", nil)
	}
	if err := validateMaterialComposition(ctx, s.repo, material.ProductID, []int64{material.MaterialID}); err != nil {
		return nil, fmt.Errorf("failed to validate material composition: %w", err)
	}

	insertedMaterial, err := s.repo.InsertMaterial(ctx, material)
	if err != nil {
		return nil, fmt.Errorf("failed to insert material: %w", err)
	}

	if insertedMaterial == nil {
		return nil, fmt.Errorf("failed to insert material")
	}

	return insertedMaterial, nil
}

func (s *MaterialService) UpdateMaterial(ctx context.Context, material entities.Material) (*entities.Material, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if material.ProductID <= 0 || material.MaterialID <= 0 {
		return nil, apperrors.BadRequest("invalid product id or material id", nil)
	}

	if !validMaterialAmount(material.MaterialAmount) {
		return nil, apperrors.BadRequest("material_amount must be a positive finite number", nil)
	}

	m, err := s.repo.UpdateMaterial(ctx, material)
	if err != nil {
		return nil, fmt.Errorf("failed to update material: %w", err)
	}

	return m, nil
}

func (s *MaterialService) DeleteMaterial(ctx context.Context, productID, materialID int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if productID <= 0 || materialID <= 0 {
		return apperrors.BadRequest("invalid product id or material id", nil)
	}

	err := s.repo.DeleteMaterial(ctx, productID, materialID)
	if err != nil {
		return fmt.Errorf("failed to delete material: %w", err)
	}

	return nil
}

func (s *MaterialService) GetMaterialsByProduct(ctx context.Context, productID int64) ([]entities.MaterialProduct, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if productID <= 0 {
		return nil, apperrors.BadRequest("invalid product id", nil)
	}

	ancestors := make(map[int64]bool)
	cache := make(map[int64][]entities.MaterialProduct)
	var buildTree func(int64) ([]entities.MaterialProduct, error)
	buildTree = func(id int64) ([]entities.MaterialProduct, error) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("request cancelled: %w", err)
		}
		if ancestors[id] {
			return nil, apperrors.Validation(fmt.Sprintf("material composition contains a cycle at product %d", id), nil)
		}
		ancestors[id] = true
		defer delete(ancestors, id)

		rows, ok := cache[id]
		if !ok {
			var err error
			rows, err = s.repo.GetMaterialsByProduct(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("failed to get materials for product %d: %w", id, err)
			}
			cache[id] = rows
		}

		materials := make([]entities.MaterialProduct, len(rows))
		for i, row := range rows {
			materialID := row.MaterialID
			if materialID == 0 {
				materialID = row.ID
			}
			children, err := buildTree(materialID)
			if err != nil {
				return nil, err
			}
			materials[i] = row
			materials[i].Materials = children
		}
		return materials, nil
	}

	tree, err := buildTree(productID)
	if err != nil {
		return nil, err
	}
	if len(tree) == 0 {
		if err := s.requireProduct(ctx, productID); err != nil {
			return nil, err
		}
	}
	return tree, nil
}

func (s *MaterialService) GetProductsByMaterial(ctx context.Context, materialID int64, limit, offset int32) ([]entities.MaterialProduct, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if materialID <= 0 {
		return nil, apperrors.BadRequest("invalid material id", nil)
	}

	products, err := s.repo.GetProductsByMaterial(ctx, materialID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by material: %w", err)
	}

	if len(products) == 0 {
		if err := s.requireProduct(ctx, materialID); err != nil {
			return nil, err
		}
	}

	return products, nil
}

// GetMaterialNorms считает сводные нормы расхода материалов на quantity единиц изделия (CALC_BOM).
func (s *MaterialService) GetMaterialNorms(ctx context.Context, productID int64, quantity float32) ([]entities.MaterialProduct, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if quantity <= 0 {
		return nil, apperrors.BadRequest("quantity must be greater than 0", nil)
	}

	rows, err := s.GetMaterialsByProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get materials by product: %w", err)
	}
	return collectMaterialNorms(rows, quantity, false), nil
}

func (s *MaterialService) requireProduct(ctx context.Context, id int64) error {
	if s.productRepo == nil {
		return fmt.Errorf("product repository is nil")
	}
	product, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get product %d: %w", id, err)
	}
	if product == nil {
		return apperrors.NotFound("product not found", nil)
	}
	return nil
}

func validMaterialAmount(amount float32) bool {
	return amount > 0 && !math.IsInf(float64(amount), 0)
}

// collectMaterialNorms группирует нормы по материалу, сохраняя каждое вхождение в состав.
func collectMaterialNorms(rows []entities.MaterialProduct, quantity float32, baseOnly bool) []entities.MaterialProduct {
	grouped := make(map[int64]entities.MaterialProduct)
	var walk func(node entities.MaterialProduct)
	walk = func(node entities.MaterialProduct) {
		if node.MaterialID == 0 {
			return
		}
		if !baseOnly || node.Base {
			// Нормы каждого вхождения масштабируются количеством конечных изделий.
			amount := node.MaterialAmount * quantity
			if existing, ok := grouped[node.MaterialID]; ok {
				existing.MaterialAmount += amount
				grouped[node.MaterialID] = existing
			} else {
				item := node
				item.MaterialAmount = amount
				item.Materials = nil
				grouped[node.MaterialID] = item
			}
		}
		for _, child := range node.Materials {
			walk(child)
		}
	}

	for _, row := range rows {
		walk(row)
	}

	result := make([]entities.MaterialProduct, 0, len(grouped))
	for _, item := range grouped {
		result = append(result, item)
	}

	return result
}
