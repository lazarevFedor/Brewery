package usecase

import (
	"Brewery/internal/entities"
	repository "Brewery/internal/repository/beer"
	"context"
)

type BeerService interface {
	CreateCategory(ctx context.Context, cat *entities.ProductCategory) (int, error)
	GetCategoryById(ctx context.Context, id int) (*entities.ProductCategory, error)
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
}

type beerService struct {
	beerRepo     repository.BeerRepository
	categoryRepo any // TODO: заменить на интерфейс CategoryRepository когда он появится
}

func NewBeerService(beerRepo repository.BeerRepository, categoryRepo any) BeerService {
	return &beerService{
		beerRepo:     beerRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *beerService) CreateCategory(ctx context.Context, cat *entities.ProductCategory) (int, error) {
	_ = ctx
	_ = cat

	return 0, nil
}

func (s *beerService) GetCategoryById(ctx context.Context, id int) (*entities.ProductCategory, error) {
	_ = ctx
	_ = id

	return nil, nil
}

func (s *beerService) UpdateCategory(ctx context.Context, id int) error {
	_ = ctx
	_ = id

	return nil
}

func (s *beerService) DeleteCategory(ctx context.Context, id int) error {
	_ = ctx
	_ = id

	return nil
}

func (s *beerService) GetAllCategories(ctx context.Context) ([]entities.ProductCategory, error) {
	_ = ctx

	return nil, nil
}

func (s *beerService) GetParentCategory(ctx context.Context, id int) (*entities.ProductCategory, error) {
	_ = ctx
	_ = id

	return nil, nil
}

func (s *beerService) GetChildCategory(ctx context.Context, id int) (*entities.ProductCategory, error) {
	_ = ctx
	_ = id

	return nil, nil
}

func (s *beerService) CreateBeer(ctx context.Context, beer *entities.Beer) (int, error) {
	_ = ctx
	_ = beer

	return 0, nil
}

func (s *beerService) UpdateBeer(ctx context.Context, id int) error {
	_ = ctx
	_ = id

	return nil
}

func (s *beerService) DeleteBeer(ctx context.Context, id int) error {
	_ = ctx
	_ = id

	return nil
}

func (s *beerService) GetAllBeers(ctx context.Context) ([]entities.Beer, error) {
	_ = ctx

	return nil, nil
}

func (s *beerService) GetBeersByCategory(ctx context.Context, id int) ([]entities.Beer, error) {
	_ = ctx
	_ = id

	return nil, nil
}
