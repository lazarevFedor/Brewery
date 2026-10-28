package usecase

import (
	"Brewery/internal/entities"
	"context"
)


type productService struct{

}


func (s *productService) CreateProduct(ctx context.Context, req entities.Product) (entities.Product, error)
