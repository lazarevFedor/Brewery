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
	GetBeerById(ctx context.Context, id int) (*entities.Beer, error)
	UpdateBeer(ctx context.Context, id int) error
	DeleteBeer(ctx context.Context, id int) error
	GetAllBeers(ctx context.Context) ([]entities.ProductCategory, error)
}

type beerService struct {
	beerRepo     *repository.BeerRepository
	categoryRepo interface{} // TODO: заменить на интерфейс CategoryRepository когда он появится
}

func NewBeerService(beerRepo *repository.BeerRepository, categoryRepo interface{}) BeerService {
	return &beerService{
		beerRepo:     beerRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *beerService) CreateCategory(ctx context.Context, cat *entities.ProductCategory) (int, error) {
	return 0, nil
}

func (s *beerService) GetCategoryById(ctx context.Context, id int) (*entities.ProductCategory, error) {
	return nil, nil
}

func (s *beerService) UpdateCategory(ctx context.Context, id int) error {
	return nil
}

func (s *beerService) DeleteCategory(ctx context.Context, id int) error {
	return nil
}

func (s *beerService) GetAllCategories(ctx context.Context) ([]entities.ProductCategory, error) {
	return nil, nil
}

func (s *beerService) GetParentCategory(ctx context.Context, id int) (*entities.ProductCategory, error) {
	return nil, nil
}

func (s *beerService) GetChildCategory(ctx context.Context, id int) (*entities.ProductCategory, error) {
	return nil, nil
}

func (s *beerService) CreateBeer(ctx context.Context, beer *entities.Beer) (int, error) {
	return 0, nil
}

func (s *beerService) GetBeerById(ctx context.Context, id int) (*entities.Beer, error) {
	return nil, nil
}

func (s *beerService) UpdateBeer(ctx context.Context, id int) error {
	return nil
}

func (s *beerService) DeleteBeer(ctx context.Context, id int) error {
	return nil
}

func (s *beerService) GetAllBeers(ctx context.Context) ([]entities.ProductCategory, error) {
	return nil, nil
}
