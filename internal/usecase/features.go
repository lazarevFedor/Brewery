package usecase

import (
	"context"
	"errors"
	"fmt"
)

func (s *beerService) GetFeatures(ctx context.Context, beerID uint) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	exists, err := s.beerRepo.BeerExists(ctx, beerID)
	if err != nil {
		return nil, err
	}
	if !exists{
		return nil, errors.New("тут надо передать ошибку 404")
	}

	features, err := s.beerRepo.GetBeerFeature(ctx, beerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return features, nil
}

func (s *beerService) CreateFeature(ctx context.Context, beerID uint, feat string) ( uint, error ) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	exists, err := s.beerRepo.BeerExists(ctx, beerID)
	if err != nil {
		return 0, err
	}
	if !exists{
		return 0, errors.New("тут надо передать ошибку 404")
	}

	featID, err := s.beerRepo.GetFeatureID(ctx, feat)
	if err != nil {
		return 0, fmt.Errorf("GetFeatureID: %w", err)
	}

	err = s.beerRepo.InsertBeerFeature(ctx, featID, beerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get categories: %w", err)
	}

	return 0, nil
}

func (s *beerService) DeleteFeature(ctx context.Context, id uint) error {
	return nil
}
