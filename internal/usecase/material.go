package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/material"
	"context"
	"errors"
	"fmt"

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

	insertedMaterial, err := s.repo.InsertMaterial(ctx, material)
	if err != nil {
		return nil, fmt.Errorf("failed to insert material: %w", err)
	}

	if insertedMaterial == nil {
		return nil, fmt.Errorf("failed to insert material")
	}

	return insertedMaterial, nil
}

func (s *MaterialService) UpdateMaterial(ctx context.Context, productID, materialID int64, amount float32) (*entities.Material, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if productID == 0 || materialID == 0 {
		return nil, errors.New("invalid product id or material id")
	}

	if amount < 0 {
		return nil, errors.New("invalid material amount")
	}

	m, err := s.repo.UpdateMaterial(ctx, entities.Material{
		ProductID:      productID,
		MaterialID:     materialID,
		MaterialAmount: amount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update material: %w", err)
	}

	return m, nil
}

func (s *MaterialService) DeleteMaterial(ctx context.Context, productID, materialID int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if productID == 0 || materialID == 0 {
		return errors.New("invalid product id or material id")
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

	if productID == 0 {
		return nil, errors.New("invalid product id")
	}

	ancestors := make(map[int64]bool)
	cache := make(map[int64][]entities.MaterialProduct)
	var buildTree func(int64) ([]entities.MaterialProduct, error)
	buildTree = func(id int64) ([]entities.MaterialProduct, error) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("request cancelled: %w", err)
		}
		if ancestors[id] {
			return nil, fmt.Errorf("material composition contains a cycle at product %d", id)
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

	return buildTree(productID)
}

func (s *MaterialService) GetProductsByMaterial(ctx context.Context, materialID int64, limit, offset int32) ([]entities.MaterialProduct, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if materialID == 0 {
		return nil, errors.New("invalid material id")
	}

	products, err := s.repo.GetProductsByMaterial(ctx, materialID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by material: %w", err)
	}

	if len(products) == 0 {
		return nil, errors.New("no products found for the given material")
	}

	return products, nil
}

// GetMaterialNorms считает сводные нормы расхода материалов на quantity единиц изделия (CALC_BOM).
func (s *MaterialService) GetMaterialNorms(ctx context.Context, productID int64, quantity float32) ([]entities.MaterialProduct, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if productID == 0 {
		return nil, errors.New("invalid product id")
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	rows, err := s.GetMaterialsByProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get materials by product: %w", err)
	}

	grouped := make(map[int64]entities.MaterialProduct)
	var walk func(node entities.MaterialProduct, ancestors map[int64]struct{})
	walk = func(node entities.MaterialProduct, ancestors map[int64]struct{}) {
		if node.MaterialID == 0 {
			return
		}
		if _, seen := ancestors[node.MaterialID]; seen {
			return
		}

		// Нормы каждого вхождения масштабируются только количеством конечных изделий.
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

		nextAncestors := make(map[int64]struct{}, len(ancestors)+1)
		for id := range ancestors {
			nextAncestors[id] = struct{}{}
		}
		nextAncestors[node.MaterialID] = struct{}{}

		for _, child := range node.Materials {
			walk(child, nextAncestors)
		}
	}

	for _, row := range rows {
		walk(row, map[int64]struct{}{})
	}

	result := make([]entities.MaterialProduct, 0, len(grouped))
	for _, item := range grouped {
		result = append(result, item)
	}

	return result, nil
}
