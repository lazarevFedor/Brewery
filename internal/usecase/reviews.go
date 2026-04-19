package usecase

import (
	"Brewery/internal/entities"
	"context"
	"fmt"
)

func (s *beerService) GetBeerReviews(ctx context.Context) ([]entities.ProductCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	categories, err := s.categoryRepo.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (s *beerService) CreateReview(ctx context.Context, review *entities.Review) (uint, error) {
	// if err := ctx.Err(); err != nil {
	// 	return 0, fmt.Errorf("request cancelled: %w", err)
	// }

	// if review == nil {
	// 	return 0, errors.New("review is nil")
	// }

	// if review.BeerID == 0 {
	// 	return 0, errors.New("invalid beer id")
	// }

	// id, err := s.beerRepo.InsertReview(ctx, *review)
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to create review: %w", err)
	// }

	return 0, nil
}

func (s *beerService) UpdateReview(ctx context.Context, id uint, updates map[string]any) (uint, error) {
	return 0, nil
}

func (s *beerService) DeleteReview(ctx context.Context, id uint) error {
	// if err := ctx.Err(); err != nil {
	// 	return fmt.Errorf("request cancelled: %w", err)
	// }

	// if id == 0 {
	// 	return errors.New("invalid beer id")
	// }

	// err := s.beerRepo.DeleteBeer(ctx, id)
	// if err != nil {
	// 	return fmt.Errorf("failed to delete beer: %w", err)
	// }

	return nil
}
