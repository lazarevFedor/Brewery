package usecase

import (
	"context"
	"fmt"
)

func (s *beerService) GetFeatures(ctx context.Context, beerID uint) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	features, err := s.beerRepo.GetBeerFeature(ctx, beerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return features, nil
}

func (s *beerService) CreateFeature(ctx context.Context, id uint, feat string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}

	featID, err := s.beerRepo.GetFeatureID(ctx, feat)
	if err != nil {
		return fmt.Errorf("GetFeatureID: %w", err)
	}

	err = s.beerRepo.InsertBeerFeature(ctx, featID, id)
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	return nil
}

func (s *beerService) DeleteFeature(ctx context.Context, id uint) error {
	return nil
}
