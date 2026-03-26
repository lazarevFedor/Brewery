package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"fmt"
)

type BeerService interface {
	CreateCategory(ctx context.Context, cat *entities.ProductCategory) (int, error)
	GetCategoryByID(ctx context.Context, id int) (*entities.ProductCategory, error)
	UpdateCategory(ctx context.Context, id int) error
	DeleteCategory(ctx context.Context, id int) error
	GetAllCategories(ctx context.Context) ([]entities.ProductCategory, error)

	GetParentCategory(ctx context.Context, id int) (*entities.ProductCategory, error)
	GetChildCategory(ctx context.Context, id int) (*entities.ProductCategory, error)

	CreateBeer(ctx context.Context, beer *entities.Beer) (int, error)
	GetBeersByCategory(ctx context.Context, id int) ([]entities.Beer, error)
	UpdateBeer(ctx context.Context, id int) error
	DeleteBeer(ctx context.Context, id int) error
	GetAllBeers(ctx context.Context) ([]entities.Beer, error)
	CreateBeerReview(ctx context.Context, id int) error
}

type beerService struct {
	beerRepo     repository.BeerRepository
	categoryRepo repository.CategoryRepository
}

func NewBeerService(beerRepo repository.BeerRepository, categoryRepo repository.CategoryRepository) BeerService {
	return &beerService{
		beerRepo:     beerRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *beerService) CreateCategory(ctx context.Context, category *entities.ProductCategory) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if category == nil {
		return 0, fmt.Errorf("category is nil")
	}

	if category.Name == "" {
		return 0, fmt.Errorf("category name is required")
	}

	id, err := s.categoryRepo.InsertCategory(ctx, *category)
	if err != nil {
		return 0, fmt.Errorf("failed to create category: %w", err)
	}

	return id, nil
}

func (s *beerService) GetCategoryByID(ctx context.Context, id int) (*entities.ProductCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return nil, fmt.Errorf("invalid category id")
	}

	ctg, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return ctg, nil
}

func (s *beerService) UpdateCategory(ctx context.Context, id int, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return fmt.Errorf("invalid category id")
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	err := s.categoryRepo.UpdateCategory(ctx, id, updates)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

func (s *beerService) DeleteCategory(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return fmt.Errorf("invalid category id")
	}

	err := s.categoryRepo.DeleteCategoryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

func (s *beerService) GetAllCategories(ctx context.Context) ([]entities.ProductCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	categories, err := s.categoryRepo.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (s *beerService) GetParentCategory(ctx context.Context, id int) (*entities.ProductCategory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *beerService) GetChildCategory(ctx context.Context, id int) (*entities.ProductCategory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *beerService) CreateBeer(ctx context.Context, beer *entities.Beer) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if beer == nil {
		return 0, fmt.Errorf("beer is nil")
	}

	if beer.Name == "" {
		return 0, fmt.Errorf("beer name is required")
	}

	if beer.Category.Name == "" {
		return 0, fmt.Errorf("category name is required")
	}

	id, err := s.beerRepo.InsertBeer(ctx, *beer)
	if err != nil {
		return 0, fmt.Errorf("failed to create beer: %w", err)
	}

	return id, nil
}

func (s *beerService) GetAllBeers(ctx context.Context) ([]entities.Beer, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	beers, err := s.beerRepo.GetBeers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get beers: %w", err)
	}

	return beers, nil
}

func (s *beerService) GetBeersByCategory(ctx context.Context, id int) ([]entities.Beer, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id <= 0 {
		return nil, fmt.Errorf("invalid category id")
	}

	beers, err := s.beerRepo.GetBeers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get beers: %w", err)
	}

	result := make([]entities.Beer, 0)

	for _, b := range beers {
		if b.Category.ID == id {
			result = append(result, b)
		}
	}

	return result, nil
}

func (s *beerService) UpdateBeer(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}

func (s *beerService) DeleteBeer(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}

func (s *beerService) CreateBeerReview(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
