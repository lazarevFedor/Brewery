package usecase

import (
	"Brewery/internal/entities"
	"context"
)

type productRepository interface {
	InsertProduct(ctx context.Context, product entities.Product) (entities.Product, error)
	GetProducts(ctx context.Context, limit, offset uint64) ([]entities.Product, error)
	GetProductByID(ctx context.Context, productID uint) (entities.Product, error)
	UpdateProduct(ctx context.Context, updates map[string]any) (entities.Product, error)
	DeleteProduct(ctx context.Context, productID uint) error
	GetProductsByClass(ctx context.Context, classID uint, limit, offset uint64) (entities.Product, error)
	GetFilteredProducts(ctx context.Context, filters []entities.FilterParameter, classID *uint, limit, offset uint64) ([]entities.Product, error)
}

type productService struct{}

func (s *productService) CreateProduct(ctx context.Context, req entities.Product) (entities.Product, error)

