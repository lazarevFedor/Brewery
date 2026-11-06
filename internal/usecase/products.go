package usecase

import (
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context, product entities.Product) (*entities.Product, error)
	GetProducts(ctx context.Context, limit, offset uint64) ([]entities.Product, error)
	GetProductByID(ctx context.Context, productID uint) (*entities.Product, error)
	UpdateProduct(ctx context.Context, productID uint, updates map[string]any) (*entities.Product, error)
	DeleteProduct(ctx context.Context, productID uint) error
	GetProductsByClass(ctx context.Context, classID uint, limit, offset uint64) ([]entities.Product, error)
	GetFilteredProducts(ctx context.Context, filters []entities.FilterParameter, classID *uint, limit, offset uint64) ([]entities.Product, error)
}

type ProductService struct {
	repo ProductRepository
}

func (ps *ProductService) CreateProduct(ctx context.Context, product entities.Product) (*entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if product.Name == "" {
		return nil, errors.New("product name is required")
	}

	//TODO: Add params' validation

	createdProduct, err := ps.repo.InsertProduct(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return createdProduct, nil
}

func (ps *ProductService) GetProducts(ctx context.Context, limit, offset uint64) ([]entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	products, err := ps.repo.GetProducts(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return products, nil
}

func (ps *ProductService) UpdateProduct(ctx context.Context, id uint, updates map[string]any) (*entities.Product, error) {
	// TODO: Validate updates
	// validatedUpdates, err := s.validateUpdates(ctx, updates)
	/* if err != nil {
		return nil, fmt.Errorf("failed to validate updates: %w", err)
	} */

	product, err := ps.repo.UpdateProduct(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (ps *ProductService) DeleteProduct(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid product id")
	}

	err := ps.repo.DeleteProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (ps *ProductService) GetProductsByClass(ctx context.Context, classID uint, limit, offset uint64) ([]entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if classID == 0 {
		return nil, errors.New("invalid class id")
	}

	products, err := ps.repo.GetProductsByClass(ctx, classID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by category: %w", err)
	}

	return products, nil
}

func (ps *ProductService) GetFilteredProducts(ctx context.Context, filters []entities.FilterParameter, classID *uint, limit, offset uint64) ([]entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	products, err := ps.repo.GetFilteredProducts(ctx, filters, classID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return products, nil
}

func (ps *ProductService) GetProductByID(ctx context.Context, id uint) (*entities.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	product, err := ps.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}
