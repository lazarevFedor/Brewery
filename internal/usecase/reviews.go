package usecase

import (
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"
)

func (s *beerService) GetBeerReviews(ctx context.Context, limit, offset uint64, beerid uint) ([]entities.Review, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	categories, err := s.beerRepo.GetReviews(ctx, limit, offset, beerid)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (s *beerService) CreateReview(ctx context.Context, review *entities.Review) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	if review.Rating >= 0 && review.Rating <= 5 {
		return 0, errors.New("review rating is out of range")
	}

	if review.BeerID == 0 {
		return 0, errors.New("invalid beer id")
	}

	id, err := s.beerRepo.InsertReview(ctx, *review)
	if err != nil {
		return 0, fmt.Errorf("failed to create review: %w", err)
	}

	return id, nil
}

func (s *beerService) UpdateReview(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	rating, ok := updates["rating"]
	if ok {
		ratingFloat, ok := rating.(float32)
		if !ok{
			return errors.New("invalid review rating datatype")
		}
		if ratingFloat >= 0 && ratingFloat <= 5 {
			return errors.New("review rating is out of range")
		}
	}

	if id == 0 {
		return errors.New("invalid beer id")
	}

	err := s.beerRepo.UpdateReview(ctx, id, updates)

	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

func (s *beerService) DeleteReview(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid beer id")
	}

	err := s.beerRepo.DeleteReview(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	return nil
}
