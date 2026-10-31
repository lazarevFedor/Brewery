// Package usecase содержит основные операции с сущностями
package usecase

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AggregateRepository определяет интерфейс для работы с агрегатами в базе данных.
type AggregateRepository interface {
	// InsertAggregate вставляет новый агрегат в базу данных и возвращает его с присвоенным ID.
	InsertAggregate(ctx context.Context, aggregate *entities.Aggregate) (*entities.Aggregate, error)

	// GetAggregates извлекает все агрегаты, соответствующие заданному имени, если оно задано, из базы данных.
	GetAggregates(ctx context.Context, name string) ([]entities.Aggregate, error)

	// ApplyAggregate применяет агрегат к категории и возвращает результат применения.
	ApplyAggregate(ctx context.Context, categoryID uint, id uint) (int, error)

	// UpdateAggregate обновляет существующий агрегат с заданным ID и возвращает обновленный агрегат.
	UpdateAggregate(ctx context.Context, id uint, updates map[string]any) (*entities.Aggregate, error)

	// DeleteAggregate удаляет агрегат с заданным ID из базы данных и возвращает удаленный агрегат.
	DeleteAggregate(ctx context.Context, id uint) (*entities.Aggregate, error)
}

// NewAggregateRepository создает новый экземпляр AggregateRepository с использованием предоставленного пула соединений к базе данных.
func NewAggregateRepository(pool *pgxpool.Pool) AggregateRepository {
	return &repository.AggregatePostgres{Pool: pool}
}

// AggregateService реализует интерфейс AggregateService.
type AggregateService struct {
	AggregateRepo AggregateRepository
}

// CreateAggregate создает новый агрегат.
func (s *AggregateService) CreateAggregate(ctx context.Context, aggregate *entities.Aggregate) (*entities.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if aggregate.Name == "" {
		return nil, errors.New("aggregate name is required")
	}

	created, err := s.AggregateRepo.InsertAggregate(ctx, aggregate)
	if err != nil {
		return nil, fmt.Errorf("failed to create aggregate: %w", err)
	}

	return created, nil
}

// GetAggregates получает список агрегатов, фильтруя по имени.
func (s *AggregateService) GetAggregates(ctx context.Context, name string) ([]entities.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	aggregates, err := s.AggregateRepo.GetAggregates(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregates: %w", err)
	}

	return aggregates, nil
}

// UpdateAggregate обновляет существующий агрегат по его ID.
func (s *AggregateService) UpdateAggregate(ctx context.Context, id uint, updates map[string]any) (*entities.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	updated, err := s.AggregateRepo.UpdateAggregate(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update aggregate: %w", err)
	}

	return updated, nil
}

// DeleteAggregate удаляет агрегат по его ID.
func (s *AggregateService) DeleteAggregate(ctx context.Context, id uint) (*entities.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request cancelled: %w", err)
	}

	deleted, err := s.AggregateRepo.DeleteAggregate(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete aggregate: %w", err)
	}

	return deleted, nil
}

// ApplyAggregateToCategory применяет агрегат к категории, возвращая количество добавленных связей.
func (s *AggregateService) ApplyAggregateToCategory(ctx context.Context, categoryID, aggregateID uint) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("request cancelled: %w", err)
	}

	added, err := s.AggregateRepo.ApplyAggregate(ctx, categoryID, aggregateID)
	if err != nil {
		return 0, fmt.Errorf("failed to apply aggregate to category: %w", err)
	}

	return added, nil
}
