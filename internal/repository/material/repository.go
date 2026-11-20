package material

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	mpr "Brewery/internal/repository/mapper"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		Base:           material.Base,
	}
	dbMaterial, err := repo.query.InsertMaterial(ctx, params)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch {
			case pgErr.Code == "23503" && pgErr.ConstraintName == "materials_product_id_fkey":
				return nil, apperrors.NotFound("product not found", err)
			case pgErr.Code == "23503" && pgErr.ConstraintName == "materials_material_id_fkey":
				return nil, apperrors.NotFound("material product not found", err)
			case pgErr.Code == "23505" && pgErr.ConstraintName == "materials_pkey":
				return nil, apperrors.Conflict("material is already part of the product", err)
			case pgErr.Code == "23514" && pgErr.ConstraintName == "chk_material_isnot_product":
				return nil, apperrors.BadRequest("product cannot be a material of itself", err)
			case pgErr.Code == "22003":
				return nil, apperrors.BadRequest("material_amount is out of range", err)
			}
		}
		return nil, fmt.Errorf("InsertMaterial: %w", err)
	}
	material = mapper.FromDBMaterial(dbMaterial)
	return &material, nil
}

func (repo *Repository) UpdateMaterial(ctx context.Context, material entities.Material) (*entities.Material, error) {
	if repo.Pool == nil {
		return nil, errors.New("db pool is empty")
	}

	updateParams := queries.UpdateMaterialParams{
		MaterialID:     material.MaterialID,
		ProductID:      material.ProductID,
		MaterialAmount: material.MaterialAmount,
		Base:           material.Base,
	}

	dbMaterial, err := repo.query.UpdateMaterial(ctx, updateParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("material not found in the product", err)
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "22003" {
			return nil, apperrors.BadRequest("material_amount is out of range", err)
		}
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
	deleted, err := repo.query.DeleteMaterial(ctx, params)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.ConstraintName == "fk_operations_material" {
			if pgErr.Code == "23503" || pgErr.Code == "23001" {
				return apperrors.Conflict("material is used by an operation", err)
			}
		}
		return fmt.Errorf("DeleteMaterial: %w", err)
	}
	if deleted == 0 {
		return apperrors.NotFound("material not found in the product", nil)
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
	dbMatProducts, err := repo.query.GetProductsByMaterial(ctx, queries.GetProductsByMaterialParams{
		MaterialID:  materialID,
		LimitCount:  limit,
		OffsetCount: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("GetProductsByMaterial: %w", err)
	}
	matProducts := mapper.FromGetByMaterial(dbMatProducts)
	return matProducts, nil
}
