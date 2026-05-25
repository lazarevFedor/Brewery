// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	if a.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	query := queries.InsertAggregate(aggregate)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build InsertAggregate query: %w", err))
	}

	var id uint
	if err = a.Pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute InsertAggregate query: %w", err))
	}

	aggregate.ID = id
	return aggregate, nil
}

// GetAggregates извлекает все агрегаты, соответствующие заданному имени, если оно задано, из базы данных.
func (a *AggregatePostgres) GetAggregates(ctx context.Context, name string) ([]entities.Aggregate, error) {
	if a.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	query := queries.GetAggregates(name)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build GetAggregates query: %w", err))
	}

	rows, err := a.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute GetAggregates query: %w", err))
	}
	defer rows.Close()

	var res []entities.Aggregate
	for rows.Next() {
		var agg entities.Aggregate
		if err = rows.Scan(&agg.ID, &agg.Name, &agg.Description, &agg.NumericParameters, &agg.EnumParameters); err != nil {
			return nil, apperrors.Internal(fmt.Errorf("scan row into aggregate: %w", err))
		}
		res = append(res, agg)
	}
	if err = rows.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("iterate over aggregate's rows: %w", err))
	}

	return res, nil
}

// ApplyAggregate применяет агрегат к категории и возвращает результат применения.
func (a *AggregatePostgres) ApplyAggregate(ctx context.Context, categoryID uint, id uint) (int, error) {
	if a.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	tx, err := a.Pool.Begin(ctx)
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("begin ApplyAggregate transaction: %w", err))
	}
	defer func() { _ = tx.Rollback(ctx) }()

	applyQuery := queries.ApplyAggregate(categoryID, id)
	sql, args, err := applyQuery.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("build apply aggregate query: %w", err))
	}

	cmd, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("execute apply aggregate query: %w", err))
	}
	affected := int(cmd.RowsAffected())

	inheritQuery := queries.InheritParametersToChildren(int(categoryID))
	sql, args, err = inheritQuery.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("build inherit parameters to children query: %w", err))
	}

	var inheritAffected int
	if err = tx.QueryRow(ctx, sql, args...).Scan(&inheritAffected); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("execute inherit parameters to children query: %w", err))
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, apperrors.Internal(fmt.Errorf("commit ApplyAggregate transaction: %w", err))
	}

	return affected + inheritAffected, nil
}

// UpdateAggregate обновляет существующий агрегат с заданным ID и возвращает обновленный агрегат.
func (a *AggregatePostgres) UpdateAggregate(ctx context.Context, id uint, updates map[string]any) (*entities.Aggregate, error) {
	if a.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	if len(updates) == 0 {
		return nil, apperrors.BadRequest("aggregate's updates doesn't provided",
			errors.New("no updates provided"))
	}

	query := queries.UpdateAggregate(id, updates)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build UpdateAggregate query: %w", err))
	}

	aggPtr, err := scanAggregate(a.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute UpdateAggregate query and scan result: %w", err))
	}
	return aggPtr, nil
}

// DeleteAggregate удаляет агрегат с заданным ID из базы данных и возвращает удаленный агрегат.
func (a *AggregatePostgres) DeleteAggregate(ctx context.Context, id uint) (*entities.Aggregate, error) {
	if a.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	query := queries.DeleteAggregate(id)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build DeleteAggregate query: %w", err))
	}

	agg, err := scanAggregate(a.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute DeleteAggregate query and scan result: %w", err))
	}
	return agg, nil
}

// scanAggregate сканирует строку результата запроса в структуру Aggregate и возвращает указатель на нее.
func scanAggregate(row pgx.Row) (*entities.Aggregate, error) {
	var agg entities.Aggregate
	if err := row.Scan(&agg.ID, &agg.Name, &agg.Description, &agg.NumericParameters, &agg.EnumParameters); err != nil {
		return nil, err
	}
	return &agg, nil
}
