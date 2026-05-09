// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/entities"
	"context"

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

// AggregatePostgres реализует интерфейс AggregateRepository для работы с агрегатами в базе данных PostgreSQL.
type AggregatePostgres struct {
	Pool *pgxpool.Pool
}

// NewAggregateRepository создает новый экземпляр AggregateRepository с использованием предоставленного пула соединений к базе данных.
func NewAggregateRepository(pool *pgxpool.Pool) AggregateRepository {
	return &AggregatePostgres{Pool: pool}
}

// NewAggregatePostgres создает новый экземпляр AggregatePostgres с использованием предоставленного пула соединений к базе данных.
func NewAggregatePostgres(pool *pgxpool.Pool) *AggregatePostgres {
	return &AggregatePostgres{Pool: pool}
}

// InsertAggregate вставляет новый агрегат в базу данных и возвращает его с присвоенным ID.
func (a *AggregatePostgres) InsertAggregate(ctx context.Context, aggregate *entities.Aggregate) (*entities.Aggregate, error) {
	return nil, nil
}

// GetAggregates извлекает все агрегаты, соответствующие заданному имени, если оно задано, из базы данных.
func (a *AggregatePostgres) GetAggregates(ctx context.Context, name string) ([]entities.Aggregate, error) {
	return nil, nil
}

// ApplyAggregate применяет агрегат к категории и возвращает результат применения.
func (a *AggregatePostgres) ApplyAggregate(ctx context.Context, categoryID uint, id uint) (int, error) {
	return 0, nil
}

// UpdateAggregate обновляет существующий агрегат с заданным ID и возвращает обновленный агрегат.
func (a *AggregatePostgres) UpdateAggregate(ctx context.Context, id uint, updates map[string]any) (*entities.Aggregate, error) {
	return nil, nil
}

// DeleteAggregate удаляет агрегат с заданным ID из базы данных и возвращает удаленный агрегат.
func (a *AggregatePostgres) DeleteAggregate(ctx context.Context, id uint) (*entities.Aggregate, error) {
	return nil, nil
}
