package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnumRepository interface {
	InsertEnumClass(ctx context.Context, enumClass entities.EnumClass) (uint, error)
	UpdateEnumClass(ctx context.Context, id uint, updates map[string]any) error
	DeleteEnumClassByID(ctx context.Context, id uint) error
	GetEnumClasses(ctx context.Context, entity, field string) ([]entities.EnumClass, error)

	InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error)
	UpdateEnumValue(ctx context.Context, id uint, value any, position *int) error
	DeleteEnumValueByID(ctx context.Context, id uint) error
	GetEnumValues(ctx context.Context, entity, field, valueType string) ([]entities.EnumValue, error)
}

type EnumPostgres struct {
	Pool *pgxpool.Pool
}

func NewEnumPostgres(pool *pgxpool.Pool) *EnumPostgres {
	return &EnumPostgres{Pool: pool}
}

func (e *EnumPostgres) InsertEnumClass(ctx context.Context, enumClass entities.EnumClass) (uint, error) {
	if e.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	var enumID uint

	psql := queries.InsertEnumClass(enumClass)

	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	err = e.Pool.QueryRow(ctx, query, args...).Scan(&enumID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "QueryRow", err)
	}

	return 0, nil
}

func (e *EnumPostgres) UpdateEnumClass(ctx context.Context, id uint, updates map[string]any) error {
	if e.Pool == nil {
		return errors.New("pool is nil")
	}

	psql := queries.UpdateEnumClass(id, updates)

	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	_, err = e.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "Exec", err)
	}

	return nil
}

func (e *EnumPostgres) DeleteEnumClassByID(ctx context.Context, id uint) error {
	if e.Pool == nil {
		return errors.New("pool is nil")
	}

	psql := queries.DeleteEnumClass(id)

	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	_, err = e.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", "Exec", err)
	}

	return nil
}

func (e *EnumPostgres) GetEnumClasses(ctx context.Context, entity, field string) ([]entities.EnumClass, error) {
	if e.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectEnumClasses(entity, field)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	rows, err := e.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Query", err)
	}
	defer rows.Close()

	enumClasses := make([]entities.EnumClass, 0)

	for rows.Next() {
		class, err := scanEnumClass(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		enumClasses = append(enumClasses, *class)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return enumClasses, nil
}

func scanEnumClass(row pgx.Row) (*entities.EnumClass, error) {
	var class entities.EnumClass

	err := row.Scan(&class.ID, &class.Type, &class.EntityName, &class.FieldName, &class.Unit, &class.IsActive)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}
	return &class, nil
}

func (e *EnumPostgres) InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error) {
	return 0, nil
}

func (e *EnumPostgres) UpdateEnumValue(ctx context.Context, id uint, value any, position *int) error {
	return nil
}

func (e *EnumPostgres) DeleteEnumValueByID(ctx context.Context, id uint) error {
	return nil
}

func (e *EnumPostgres) GetEnumValues(ctx context.Context, entity, field, valueType string) ([]entities.EnumValue, error) {
	return nil, nil
}
