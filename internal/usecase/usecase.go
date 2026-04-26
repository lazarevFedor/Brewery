// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"
)

type BeerService interface {
	CreateCategory(ctx context.Context, ctg *entities.ProductCategory) (uint, error)
	GetCategoryByID(ctx context.Context, id uint) (*entities.ProductCategory, error)
	UpdateCategory(ctx context.Context, id uint, updates map[string]any) error
	DeleteCategory(ctx context.Context, id uint) error
	GetAllCategories(ctx context.Context) ([]entities.ProductCategory, error)

	GetParentCategory(ctx context.Context, id uint) (*entities.ProductCategory, error)
	GetChildCategories(ctx context.Context, id uint) ([]entities.ProductCategory, error)

	CreateBeer(ctx context.Context, beer *entities.Beer) (*entities.Beer, error)
	GetBeersByCategory(ctx context.Context, id uint, limit, offset uint64) ([]entities.Beer, error)
	UpdateBeer(ctx context.Context, id uint, updates map[string]any) (uint, error)
	DeleteBeer(ctx context.Context, id uint) error
	GetAllBeers(ctx context.Context, limit, offset uint64) ([]entities.Beer, error)

	GetFeatures(ctx context.Context, id uint) ([]string, error)
	CreateFeature(ctx context.Context, id uint, feat string) error
	DeleteFeature(ctx context.Context, id uint) error

	GetBeerReviews(ctx context.Context, limit, offset uint64, beerid uint) ([]entities.Review, error)
	CreateReview(ctx context.Context, review *entities.Review) (uint, error)
	UpdateReview(ctx context.Context, id uint, updates map[string]any) error
	DeleteReview(ctx context.Context, id uint) error
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

func (s *beerService) CreateCategory(ctx context.Context, ctg *entities.ProductCategory) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if ctg == nil {
		return 0, errors.New("category is nil")
	}

	if ctg.Name == "" {
		return 0, errors.New("category name is required")
	}

	if ctg.ParentID == 0 {
		categories, err := s.categoryRepo.GetCategories(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to check existing categories: %w", err)
		}
		for _, c := range categories {
			if c.ParentID == 0 {
				return 0, errors.New("root category already exists")
			}
		}
	} else {
		parent, err := s.categoryRepo.GetCategoryByID(ctx, uint(ctg.ParentID))
		if err != nil {
			return 0, fmt.Errorf("failed to get parent category: %w", err)
		}
		if parent == nil {
			return 0, errors.New("parent category not found")
		}
	}

	id, err := s.categoryRepo.InsertCategory(ctx, *ctg)
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
		return nil, errors.New("invalid category id")
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
		return errors.New("invalid category id")
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	parentID, ok := updates["parent_id"]
	if ok {
		parentIDFloat, ok := parentID.(float64)
		if ok {
			if parentIDFloat == 0 {
				updates["parent_id"] = nil
			} else {
				err := s.ensureCategoryParentIsNotDescendant(ctx, id, uint(parentIDFloat))
				if err != nil {
					return fmt.Errorf("failed to update parent category: %w", err)
				}
			}
		}
	}

	err := s.categoryRepo.UpdateCategory(ctx, id, updates)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

func (s *beerService) ensureCategoryParentIsNotDescendant(ctx context.Context, categoryID, parentID uint) error {
	visited := make(map[uint]struct{})
	currentID := parentID

	for currentID != 0 {
		if _, seen := visited[currentID]; seen {
			return errors.New("category hierarchy contains a cycle")
		}
		visited[currentID] = struct{}{}

		category, err := s.categoryRepo.GetCategoryByID(ctx, currentID)
		if err != nil {
			return fmt.Errorf("failed to get parent category: %w", err)
		}
		if category == nil {
			return errors.New("parent category not found")
		}

		if category.ID == int(categoryID) {
			return errors.New("category parent creates a cycle")
		}

		currentID = uint(category.ParentID)
	}

	return nil
}

func (s *beerService) DeleteCategory(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid category id")
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
		return nil, nil
	}

	parent, err := s.categoryRepo.GetCategoryByID(ctx, uint(ctg.ParentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get parent category: %w", err)
	}

	return parent, nil
}

func (s *beerService) GetChildCategories(ctx context.Context, id uint) ([]entities.ProductCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	categories, err := s.categoryRepo.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	children := make([]entities.ProductCategory, 0)
	for _, c := range categories {
		if c.ParentID != 0 && uint(c.ParentID) == id {
			children = append(children, c)
		}
	}
	return children, nil
}

func (s *beerService) CreateBeer(ctx context.Context, beer *entities.Beer) (*entities.Beer, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if beer == nil {
		return nil, errors.New("beer is nil")
	}

	if beer.Name == "" {
		return nil, errors.New("beer name is required")
	}

	beer, err := s.beerRepo.InsertBeer(ctx, *beer)
	if err != nil {
		return nil, fmt.Errorf("failed to create beer: %w", err)
	}

	return beer, nil
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
		return nil, errors.New("invalid category id")
	}

	beers, err := s.beerRepo.GetBeersByCategoryID(ctx, id, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get beers by category: %w", err)
	}

	return beers, nil
}

func validation(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid beer id")
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}
	return nil
}

func (s *beerService) resolveCityUpdate(ctx context.Context, updates map[string]any) (uint, error) {
	countryName, ok := updates["country"].(string)
	if !ok {
		return 0, errors.New("country Datatype error")
	}
	ctrID, err := s.beerRepo.GetCountryID(ctx, countryName)
	if err != nil {
		return 0, fmt.Errorf("failed to get Country ID: %w", err)
	}

	city := updates["city"]
	cityName, ok := city.(string)
	if !ok {
		return 0, errors.New("cityName Datatype error")
	}
	cityID, err := s.beerRepo.GetCityID(ctx, cityName, ctrID)
	if err != nil {
		return 0, err
	}

	return cityID, nil
}

func (s *beerService) UpdateBeer(ctx context.Context, id uint, updates map[string]any) (uint, error) {
	if err := validation(ctx, id, updates); err != nil {
		return 0, err
	}
	finalUpdates := make(map[string]any)

	for k, v := range updates {
		switch k {
		case "city":
			cityID, err := s.resolveCityUpdate(ctx, updates)
			if err != nil {
				return 0, err
			}
			finalUpdates["city_id"] = cityID

		case "category":
			categoryUpdates, ok := v.(map[string]any)
			if !ok {
				return 0, errors.New("category datatype error")
			}

			updateCtgID, ok := categoryUpdates["id"]
			if ok {
				updateCtgIDFLoat, ok := updateCtgID.(float64)
				if !ok {
					return 0, errors.New("category id datatype error")
				}

				ctg, err := s.categoryRepo.GetCategoryByID(ctx, uint(updateCtgIDFLoat))
				if err != nil {
					return 0, fmt.Errorf("failed to get ctg by id: %w", err)
				}
				finalUpdates["category_id"] = uint(ctg.ID)
			} else {
				ctgName, ok := categoryUpdates["name"]
				if !ok {
					return 0, errors.New("category name needs to update category")
				}

				ctgNameStr, ok := ctgName.(string)
				if !ok {
					return 0, errors.New("category name datatype error")
				}

				ctgID, err := s.categoryRepo.GetCategoryID(ctx, ctgNameStr)
				if err != nil {
					return 0, fmt.Errorf("failed to get Category ID: %w", err)
				}
				if ctgID == 0 {
					parentID, ok := categoryUpdates["parent_id"]
					if !ok {
						return 0, errors.New("category name needs to update category")
					}

					parentIDFloat, ok := parentID.(float64)
					if !ok {
						return 0, errors.New("parent_id datatype error")
					}
					ctgID, err = s.categoryRepo.InsertCategory(ctx, entities.ProductCategory{
						Name:     ctgNameStr,
						ParentID: int(parentIDFloat),
					})
					if err != nil {
						return 0, fmt.Errorf("insertcategory: %w", err)
					}
				}
				finalUpdates["category_id"] = ctgID
			}

		default:
			if k != "country" {
				finalUpdates[k] = v
			}
		}
	}

	beerID, err := s.beerRepo.UpdateBeer(ctx, id, finalUpdates)
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
		return errors.New("invalid beer id")
	}

	err := s.beerRepo.DeleteBeer(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete beer: %w", err)
	}

	return nil
}
