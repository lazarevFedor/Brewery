package material

import (
	"Brewery/internal/entities"
	mpr "Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var mapper = mpr.MaterialMapperImpl{}

type Repository struct {
	Pool  *pgxpool.Pool
	query *queries.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Pool:  pool,
		query: queries.New(pool),
	}
}

func (repo *Repository) InsertMaterial(ctx context.Context, material entities.Material) (*entities.Material, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	params := queries.InsertMaterialParams{
		MaterialID:     material.MaterialID,
		ProductID:      material.ProductID,
		MaterialAmount: material.MaterialAmount,
	}
	dbMaterial, err := repo.query.InsertMaterial(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("InsertMaterial: %w", err)
	}
	material = mapper.FromDBMaterial(dbMaterial)
	return &material, nil
}

func (repo *Repository) UpdateMaterial(ctx context.Context, material entities.Material) (*entities.Material, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	var updateParams queries.UpdateMaterialParams
	if material.MaterialID == 0 {
		updateParams.MaterialID = nil
	}
	if material.ProductID == 0 {
		updateParams.ProductID = nil
	}
	if material.MaterialAmount == 0 {
		updateParams.MaterialAmount = nil
	}

	dbMaterial, err := repo.query.UpdateMaterial(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("UpdateMaterial: %w", err)
	}
	material = mapper.FromDBMaterial(dbMaterial)
	return &material, nil
}

func (repo *Repository) DeleteMaterial(ctx context.Context, productID, materialID int64) error {
	if repo.Pool == nil {
		return errors.New("db pool is empty")
	}
	params := queries.DeleteMaterialParams{
		MaterialID: materialID,
		ProductID:  productID,
	}
	if err := repo.query.DeleteMaterial(ctx, params); err != nil {
		return fmt.Errorf("DeleteMaterial: %w", err)
	}
	return nil
}

func (repo *Repository) GetMaterialsByProduct(ctx context.Context, productID int64) ([]entities.MaterialProduct, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	dbMaterials, err := repo.query.GetMaterialsByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("GetMaterialsByProductID: %w", err)
	}
	matProducts := mapper.FromGetByProduct(dbMaterials)
	return matProducts, nil
}

func (repo *Repository) GetProductsByMaterial(ctx context.Context, materialID int64, limit, offset int32) ([]entities.MaterialProduct, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}
	dbMatProducts, err := repo.query.GetProductsByMaterial(ctx, materialID)
	if err != nil {
		return nil, fmt.Errorf("GetProductsByMaterial: %w", err)
	}
	matProducts := mapper.FromGetByMaterial(dbMatProducts)
	return matProducts, nil
}

// CopyProductMaterials копирует материалы из базового изделия в модификацию.
func (repo *Repository) CopyProductMaterials(ctx context.Context, sourceProductID, targetProductID int64) error{
	if repo.Pool == nil {
		return errors.New("db pool is empty")
	}	
	params := queries.CopyBaseMaterialsParams{
		ProductID: targetProductID,
		BaseID: sourceProductID,
	}
	err := repo.query.CopyBaseMaterials(ctx, params)
	if err != nil {
		return fmt.Errorf("GetProductsByMaterial: %w", err)
	}
	return nil
}