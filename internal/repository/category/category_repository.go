package category

import (
	"context"
	"Brewery/internal/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
}

type CategoryPostgres struct {
	pool *pgxpool.Pool
}

func NewCategoryPostgres(pool *pgxpool.Pool) *CategoryPostgres {
	return &CategoryPostgres{pool: pool}
}

func (r *CategoryPostgres) InsertCategory(ctx context.Context, category entities.ProductCategory) error {
	return nil
}

func (r *CategoryPostgres) DeleteCategoryByName(ctx context.Context, name string) error {
	return nil
}
