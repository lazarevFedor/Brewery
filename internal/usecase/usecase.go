package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"fmt"
)

type BeerService interface {
	CreateCategory(ctx context.Context, cat *entities.ProductCategory) (uint, error)
	GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error)
	UpdateCategory(ctx context.Context, id uint, updates map[string]any) error
	DeleteCategory(ctx context.Context, id uint) error
	GetAllCategories(ctx context.Context) ([]entities.ProductCategory, error)

	GetParentCategory(ctx context.Context, id uint) (*entities.ProductCategory, error)
	GetChildCategory(ctx context.Context, id uint) (*entities.ProductCategory, error)

	CreateBeer(ctx context.Context, beer *entities.Beer) (uint, error)
	GetBeersByCategory(ctx context.Context, id uint, limit, offset uint64) ([]entities.Beer, error)
	UpdateBeer(ctx context.Context, id uint, updates map[string]any) (uint, error)
	DeleteBeer(ctx context.Context, id uint) error
	GetAllBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error)
	CreateBeerReview(ctx context.Context, review *entities.Review) (uint, error)
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

func (s *beerService) CreateCategory(ctx context.Context, cat *entities.ProductCategory) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if cat == nil {
		return 0, fmt.Errorf("category is nil")
	}

	if cat.Name == "" {
		return 0, fmt.Errorf("category name is required")
	}

	id, err := s.categoryRepo.InsertCategory(ctx, *cat)
	if err != nil {
		return 0, fmt.Errorf("failed to create category: %w", err)
	}

	return id, nil
}

func (s *beerService) GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, fmt.Errorf("invalid category id")
	}

	ctg, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return ctg, nil
}

func (s *beerService) UpdateCategory(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
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

func (s *beerService) DeleteCategory(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
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

func (s *beerService) GetParentCategory(ctx context.Context, id uint) (*entities.ProductCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	ctg, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if ctg.ParentID == 0 {
		return nil, fmt.Errorf("category has no parent")
	}

	parent, err := s.categoryRepo.GetCategoryByID(ctx, uint(ctg.ParentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get parent category: %w", err)
	}

	return parent, nil
}

func (s *beerService) GetChildCategory(ctx context.Context, id uint) (*entities.ProductCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	categories, err := s.categoryRepo.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	for _, c := range categories {
		if c.ParentID != 0 && uint(c.ParentID) == id {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("child category not found")
}

func (s *beerService) CreateBeer(ctx context.Context, beer *entities.Beer) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if beer == nil {
		return 0, fmt.Errorf("beer is nil")
	}

	if beer.Name == "" {
		return 0, fmt.Errorf("beer name is required")
	}

	id, err := s.beerRepo.InsertBeer(ctx, *beer)
	if err != nil {
		return 0, fmt.Errorf("failed to create beer: %w", err)
	}

	return id, nil
}

func (s *beerService) GetAllBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	beers, err := s.beerRepo.GetBeers(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get beers: %w", err)
	}

	return beers, nil
}

func (s *beerService) GetBeersByCategory(ctx context.Context, id uint, limit, offset uint64) ([]entities.Beer, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return nil, fmt.Errorf("invalid category id")
	}

	beers, err := s.beerRepo.GetBeersByCategoryID(ctx, id, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get beers by category: %w", err)
	}

	return beers, nil
}

func (s *beerService) UpdateBeer(ctx context.Context, id uint, updates map[string]any) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return 0, fmt.Errorf("invalid beer id")
	}

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	beerID, err := s.beerRepo.UpdateBeer(ctx, id, updates)
	if err != nil {
		return 0, fmt.Errorf("failed to update beer: %w", err)
	}

	return beerID, nil
}

func (s *beerService) DeleteBeer(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return fmt.Errorf("invalid beer id")
	}

	err := s.beerRepo.DeleteBeer(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete beer: %w", err)
	}

	return nil
}

func (s *beerService) CreateBeerReview(ctx context.Context, review *entities.Review) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if review == nil {
		return 0, fmt.Errorf("review is nil")
	}

	if review.BeerID == 0 {
		return 0, fmt.Errorf("invalid beer id")
	}

	id, err := s.beerRepo.InsertReview(ctx, *review)
	if err != nil {
		return 0, fmt.Errorf("failed to create review: %w", err)
	}

	return id, nil
}
