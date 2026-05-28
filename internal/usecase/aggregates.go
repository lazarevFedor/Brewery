// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"
)

// AggregateService определяет интерфейс для работы с агрегатами.
type AggregateService interface {

	// CreateAggregate создает новый агрегат.
	CreateAggregate(ctx context.Context, aggregate *entities.Aggregate) (*entities.Aggregate, error)

	// GetAggregates получает список агрегатов, фильтруя по имени.
	GetAggregates(ctx context.Context, name string) ([]entities.Aggregate, error)

	// UpdateAggregate обновляет существующий агрегат по его ID.
	UpdateAggregate(ctx context.Context, id uint, updates map[string]any) (*entities.Aggregate, error)

	// DeleteAggregate удаляет агрегат по его ID.
	DeleteAggregate(ctx context.Context, id uint) (*entities.Aggregate, error)

	// ApplyAggregateToCategory применяет агрегат к категории, возвращая количество добавленных связей.
	ApplyAggregateToCategory(ctx context.Context, categoryID, aggregateID uint) (int, error)
}

// aggregateService реализует интерфейс AggregateService.
type aggregateService struct {
	aggregateRepo repository.AggregateRepository
}

// NewAggregateService создает новый экземпляр AggregateService.
func NewAggregateService(aggregateRepo repository.AggregateRepository) AggregateService {
	return &aggregateService{
		aggregateRepo: aggregateRepo,
	}
}

// CreateAggregate создает новый агрегат.
func (s *aggregateService) CreateAggregate(ctx context.Context, aggregate *entities.Aggregate) (*entities.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	if aggregate.Name == "" {
		return nil, apperrors.BadRequest("aggregate name is required", errors.New("aggregate name is required"))
	}

	created, err := s.aggregateRepo.InsertAggregate(ctx, aggregate)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// GetAggregates получает список агрегатов, фильтруя по имени.
func (s *aggregateService) GetAggregates(ctx context.Context, name string) ([]entities.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	aggregates, err := s.aggregateRepo.GetAggregates(ctx, name)
	if err != nil {
		return nil, err
	}

	return aggregates, nil
}

// UpdateAggregate обновляет существующий агрегат по его ID.
func (s *aggregateService) UpdateAggregate(ctx context.Context, id uint, updates map[string]any) (*entities.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	if len(updates) == 0 {
		return nil, apperrors.BadRequest("no updates provided", errors.New("no updates provided"))
	}

	updated, err := s.aggregateRepo.UpdateAggregate(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// DeleteAggregate удаляет агрегат по его ID.
func (s *aggregateService) DeleteAggregate(ctx context.Context, id uint) (*entities.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	deleted, err := s.aggregateRepo.DeleteAggregate(ctx, id)
	if err != nil {
		return nil, err
	}

	return deleted, nil
}

// ApplyAggregateToCategory применяет агрегат к категории, возвращая количество добавленных связей.
func (s *aggregateService) ApplyAggregateToCategory(ctx context.Context, categoryID, aggregateID uint) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("request cancelled: %w", err))
	}

	added, err := s.aggregateRepo.ApplyAggregate(ctx, categoryID, aggregateID)
	if err != nil {
		return 0, err
	}

	return added, nil
}
