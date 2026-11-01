package repository

import (
	"Brewery/internal/entities"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MaterialPostgres реализует интерфейс MaterialRepository для работы с материалами в базе данных PostgreSQL.
type MaterialPostgres struct {
	Pool *pgxpool.Pool
}

// NewMaterialPostgres создает новый экземпляр MaterialPostgres с использованием предоставленного пула соединений к базе данных.
func NewMaterialPostgres(pool *pgxpool.Pool) *MaterialPostgres {
	return &MaterialPostgres{Pool: pool}
}

func (a *MaterialPostgres) InsertMaterial(ctx context.Context, material *entities.Material) (*entities.Material, error) {
	return nil, nil
}

func (a *MaterialPostgres) UpdateMaterial(ctx context.Context, id uint, updates map[string]any) (*entities.Material, error) {
	return nil, nil
}

func (a *MaterialPostgres) DeleteMaterial(ctx context.Context, id uint) (*entities.Material, error) {
	return nil, nil
}

func (a *MaterialPostgres) GetProductMaterials(ctx context.Context, productID uint) ([]entities.Material, error) {
	return nil, nil
}

func (a *MaterialPostgres) GetProductsUseMaterial(ctx context.Context, materialID uint) ([]entities.Product, error) {
	return nil, nil
}
