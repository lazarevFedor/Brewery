// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"errors"
	"fmt"
)

// GetBeerReviews получает список отзывов для пива.
func (s *beerService) GetBeerReviews(ctx context.Context, limit, offset uint64, beerid uint) ([]entities.Review, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	categories, err := s.beerRepo.GetReviews(ctx, limit, offset, beerid)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// CreateReview создает новый отзыв для пива.
func (s *beerService) CreateReview(ctx context.Context, review *entities.Review) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	if review.Rating < 1 || review.Rating > 5 {
		return 0, apperrors.BadRequest("invalid review rating", errors.New("review rating must be between 1 and 5"))
	}

	if review.BeerID == 0 {
		return 0, apperrors.NotFound("beer not found", errors.New("beer with the given ID does not exist"))
	}

	if review.Body == "" {
		return 0, apperrors.BadRequest("invalid review body", errors.New("review body cannot be empty"))
	}

	id, err := s.beerRepo.InsertReview(ctx, *review)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateReview обновляет существующий отзыв для пива.
func (s *beerService) UpdateReview(ctx context.Context, id uint, updates map[string]any) error {
	if err := ctx.Err(); err != nil {
		return apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	rating, ok := updates["rating"]
	if ok {
		ratingFloat, ok := rating.(float32)
		if !ok {
			return apperrors.Internal(errors.New("invalid type for review rating"))
		}
		if ratingFloat >= 0 && ratingFloat <= 5 {
			return apperrors.BadRequest("invalid review rating", errors.New("review rating must be between 0 and 5"))
		}
	}

	if id == 0 {
		return apperrors.NotFound("review not found", errors.New("review with the given ID does not exist"))
	}

	if len(updates) == 0 {
		return apperrors.BadRequest("no updates provided", errors.New("no updates provided"))
	}

	if body, ok := updates["body"]; ok {
		if bodyStr, ok := body.(string); ok && bodyStr == "" {
			return apperrors.BadRequest("invalid review body", errors.New("review body cannot be empty"))
		}
	}

	err := s.beerRepo.UpdateReview(ctx, id, updates)
	if err != nil {
		return err
	}

	return nil
}

// DeleteReview удаляет существующий отзыв для пива.
func (s *beerService) DeleteReview(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	if id == 0 {
		return apperrors.NotFound("review not found", errors.New("review with the given ID does not exist"))
	}

	err := s.beerRepo.DeleteReview(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
