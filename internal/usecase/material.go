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
	return &material.Repository{Pool: pool}
}

type MaterialService struct {
	repo        MaterialRepository
	productRepo ProductRepository
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

	materials, err := s.repo.GetMaterialsByProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get materials by product: %w", err)
	}

	return materials, nil
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

func (s *MaterialService) GetMaterialNorms(ctx context.Context, productID int64) ([]entities.MaterialProduct, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if productID == 0 {
		return nil, errors.New("invalid product id")
	}

	active := make(map[int64]struct{})
	grouped := make(map[int64]entities.MaterialProduct)

	var walk func(productID int64) error

	walk = func(productID int64) error {
		if _, ok := active[productID]; ok {
			return nil
		}
		active[productID] = struct{}{}
		defer delete(active, productID)

		rows, err := s.repo.GetMaterialsByProduct(ctx, productID)
		if err != nil {
			return fmt.Errorf("failed to get materials by product %d: %w", productID, err)
		}

		for _, row := range rows {
			if existing, ok := grouped[row.ProductID]; ok {
				existing.MaterialAmount += row.MaterialAmount
				grouped[row.ProductID] = existing
			} else {
				grouped[row.ProductID] = row
			}

			if err := walk(row.MaterialID); err != nil {
				return err
			}
		}

		return nil
	}

	if err := walk(productID); err != nil {
		return nil, err
	}

	result := make([]entities.MaterialProduct, 0, len(grouped))
	for _, v := range grouped {
		result = append(result, v)
	}

	return result, nil
}
