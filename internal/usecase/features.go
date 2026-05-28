// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/apperrors"
	"context"
	"fmt"
)

// GetFeatures возвращает список характеристик для пива по его ID.
func (s *beerService) GetFeatures(ctx context.Context, beerID uint) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	exists, err := s.beerRepo.BeerExists(ctx, beerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.NotFound("beer not found", fmt.Errorf("beer with id %d not found", beerID))
	}

	features, err := s.beerRepo.GetBeerFeature(ctx, beerID)
	if err != nil {
		return nil, err
	}

	return features, nil
}

// CreateFeature добавляет характеристику к пиву по его ID.
func (s *beerService) CreateFeature(ctx context.Context, beerID uint, feat string) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	exists, err := s.beerRepo.BeerExists(ctx, beerID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, apperrors.NotFound("beer not found", fmt.Errorf("beer with id %d not found", beerID))
	}

	featID, err := s.beerRepo.GetFeatureID(ctx, nil, feat)
	if err != nil {
		return 0, err
	}

	err = s.beerRepo.ConnectBeerAndFeature(ctx, nil, featID, beerID)
	if err != nil {
		return 0, err
	}

	return 0, nil
}

// DeleteFeature удаляет характеристику у пива по его ID.
func (s *beerService) DeleteFeature(ctx context.Context, beerID uint) error {
	if err := ctx.Err(); err != nil {
		return apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	exists, err := s.beerRepo.BeerExists(ctx, beerID)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.NotFound("beer not found", fmt.Errorf("beer with id %d not found", beerID))
	}

	err = s.beerRepo.DisconnectBeerAndFeature(ctx, nil, beerID)
	if err != nil {
		return err
	}

	return nil
}
