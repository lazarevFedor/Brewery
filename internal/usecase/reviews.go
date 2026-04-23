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

	if review == nil {
		return 0, errors.New("review is nil")
	}

	if review.BeerID == 0 {
		return 0, errors.New("invalid beer id")
	}

	if review.Rating < 0 || review.Rating > 5 {
		return 0, errors.New("rating must be between 0 and 5")
	}

	if review.Body == "" {
		return 0, errors.New("review body cannot be empty")
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

	if updates == nil {
		return errors.New("updates map is nil")
	}

	if id == 0 {
		return errors.New("invalid review id")
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	if rating, ok := updates["rating"]; ok {
		switch v := rating.(type) {
		case float64:
			if v < 0 || v > 5 {
				return errors.New("rating must be between 0 and 5")
			}
		case float32:
			if v < 0 || v > 5 {
				return errors.New("rating must be between 0 and 5")
			}
		case int:
			if v < 0 || v > 5 {
				return errors.New("rating must be between 0 and 5")
			}
		case int64:
			if v < 0 || v > 5 {
				return errors.New("rating must be between 0 and 5")
			}
		case int32:
			if v < 0 || v > 5 {
				return errors.New("rating must be between 0 and 5")
			}
		}
	}

	if body, ok := updates["body"]; ok {
		if bodyStr, ok := body.(string); ok && bodyStr == "" {
			return errors.New("review body cannot be empty")
		}
	}

	err := s.beerRepo.UpdateReview(ctx, id, updates)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	return nil
}

func (s *beerService) DeleteReview(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	if id == 0 {
		return errors.New("invalid review id")
	}

	err := s.beerRepo.DeleteReview(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	return nil
}
