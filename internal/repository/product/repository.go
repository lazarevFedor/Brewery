package product

import (
	"Brewery/internal/entities"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	Pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{Pool: pool}
}

func (repo *Repository) InsertProduct(ctx context.Context, product entities.Product) (entities.Product, error)
func (repo *Repository) GetProducts(ctx context.Context, limit, offset uint64) ([]entities.Product, error)
func (repo *Repository) GetProductByID(ctx context.Context, productID uint) (entities.Product, error)
func (repo *Repository) UpdateProduct(ctx context.Context, updates map[string]any) (entities.Product, error)
func (repo *Repository) DeleteProduct(ctx context.Context, productID uint) error
func (repo *Repository) GetProductsByClass(
	ctx context.Context, classID uint, limit, offset uint64,
) (entities.Product, error)

func (repo *Repository) GetFilteredProducts(
	ctx context.Context, filters []entities.FilterParameter,
	classID *uint, limit, offset uint64,
) ([]entities.Product, error)
