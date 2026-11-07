package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/material"
	"context"

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

	// GetProductMaterials извлекает все материалы, соответствующие заданному ID продукта, из базы данных.
	GetMaterialsByProduct(ctx context.Context, productID int64) ([]entities.MaterialProduct, error)

	//GetProductsByMaterial извлекает все продукты, использующие заданный материал, из базы данных.
	GetProductsByMaterial(ctx context.Context, materialID int64, limit, offset int32) ([]entities.MaterialProduct, error)
}

// NewMaterialRepository создает новый экземпляр MaterialRepository с использованием предоставленного пула соединений к базе данных.
func NewMaterialRepository(pool *pgxpool.Pool) MaterialRepository {
	return &material.Repository{Pool: pool}
}
