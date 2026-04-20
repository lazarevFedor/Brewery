package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnumRepository определяет контракт для взаимодействия с классами и значениями перечислений
type EnumRepository interface {

	// InsertEnumClass сохраняет новую сущность EnumClass в хранилище.
	InsertEnumClass(ctx context.Context, enumClass entities.EnumClass) (uint, error)

	// UpdateEnumClass обновляет сущность EnumClass в хранилище.
	UpdateEnumClass(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumClassByID удаляет сущность EnumClass из хранилища.
	DeleteEnumClassByID(ctx context.Context, id uint) error

	// GetEnumClasses получает список сущностей EnumClass по заданным имени таблицы и поля.
	GetEnumClasses(ctx context.Context, entity, field string) ([]entities.EnumClass, error)

	// InsertEnumValue сохраняет новую сущность EnumValue в хранилище.
	InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error)

	// UpdateEnumValue обновляет сущность EnumValue в хранилище.
	UpdateEnumValue(ctx context.Context, id uint, value any, position *int) error

	// DeleteEnumValueByID удаляет сущность EnumValue из хранилища.
	DeleteEnumValueByID(ctx context.Context, id uint) error

	// GetEnumValues получает список сущностей EnumClass по заданным имени таблицы, поля и типу значения.
	GetEnumValues(ctx context.Context, entity, field, valueType string) ([]entities.EnumValue, error)
}

// EnumPostgres хранит в себе пул подключений к БД
type EnumPostgres struct {
	Pool *pgxpool.Pool
}

// NewEnumPostgres создает новый репозиторий БД
func NewEnumPostgres(pool *pgxpool.Pool) *EnumPostgres {
	return &EnumPostgres{Pool: pool}
}

// InsertEnumClass сохраняет новую сущность EnumClass в хранилище.
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

	return enumID, nil
}

// UpdateEnumClass обновляет сущность EnumClass в хранилище.
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

// DeleteEnumClassByID удаляет сущность EnumClass из хранилища.
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

// GetEnumClasses получает список сущностей EnumClass по заданным имени таблицы и поля.
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
	var unit pgtype.Text

	err := row.Scan(&class.ID, &class.Type, &class.EntityName, &class.FieldName, &unit, &class.IsActive)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	if unit.Valid {
		class.Unit = unit.String
	} else {
		class.Unit = ""
	}

	return &class, nil
}

// InsertEnumValue сохраняет новую сущность EnumValue в хранилище.
func (e *EnumPostgres) InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error) {
	return 0, nil
}

// UpdateEnumValue обновляет сущность EnumValue в хранилище.
func (e *EnumPostgres) UpdateEnumValue(ctx context.Context, id uint, value any, position *int) error {
	return nil
}

// DeleteEnumValueByID удаляет сущность EnumValue из хранилища.
func (e *EnumPostgres) DeleteEnumValueByID(ctx context.Context, id uint) error {
	return nil
}

// GetEnumValues получает список сущностей EnumClass по заданным имени таблицы, поля и типу значения.
func (e *EnumPostgres) GetEnumValues(ctx context.Context, entity, field, valueType string) ([]entities.EnumValue, error) {
	return nil, nil
}
