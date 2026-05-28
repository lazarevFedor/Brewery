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
	UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error

	// DeleteEnumValueByID удаляет сущность EnumValue из хранилища.
	DeleteEnumValueByID(ctx context.Context, id uint) error

	// GetEnumValues получает список сущностей EnumValue по заданным имени таблицы, поля и типу значения.
	GetEnumValues(ctx context.Context, entity, field string, valueType entities.EnumType) ([]entities.EnumValue, error)

	// GetEnumValuesByClassID получает список сущностей EnumValue по заданному ID класса перечисления.
	GetEnumValuesByClassID(ctx context.Context, classID uint) ([]entities.EnumValue, error)

	// GetEnumClassByID получает сущность EnumClass по заданному ID класса перечисления.
	GetEnumClassByID(ctx context.Context, id uint) (*entities.EnumClass, error)
}

// EnumPostgres хранит в себе пул подключений к БД
type EnumPostgres struct {
	Pool *pgxpool.Pool
}

// NewEnumRepository создает новый экземпляр EnumRepository с переданным пулом соединений.
func NewEnumRepository(pool *pgxpool.Pool) EnumRepository {
	return &EnumPostgres{Pool: pool}
}

// NewEnumPostgres создает новый репозиторий БД
func NewEnumPostgres(pool *pgxpool.Pool) *EnumPostgres {
	return &EnumPostgres{Pool: pool}
}

// InsertEnumClass сохраняет новую сущность EnumClass в хранилище.
func (e *EnumPostgres) InsertEnumClass(ctx context.Context, enumClass entities.EnumClass) (uint, error) {
	if e.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	var enumID uint

	psql := queries.InsertEnumClass(enumClass)

	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("build InsertEnumClass query: %w", err))
	}

	err = e.Pool.QueryRow(ctx, query, args...).Scan(&enumID)
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("execute InsertEnumClass query: %w", err))
	}

	return enumID, nil
}

// UpdateEnumClass обновляет сущность EnumClass в хранилище.
func (e *EnumPostgres) UpdateEnumClass(ctx context.Context, id uint, updates map[string]any) error {
	if e.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.UpdateEnumClass(id, updates)

	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("build UpdateEnumClass query: %w", err))
	}

	_, err = e.Pool.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("execute UpdateEnumClass query: %w", err))
	}

	return nil
}

// DeleteEnumClassByID удаляет сущность EnumClass из хранилища.
func (e *EnumPostgres) DeleteEnumClassByID(ctx context.Context, id uint) error {
	if e.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.DeleteEnumClass(id)

	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("build DeleteEnumClassByID query: %w", err))
	}

	_, err = e.Pool.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("execute DeleteEnumClassByID query: %w", err))
	}

	return nil
}

// GetEnumClasses получает список сущностей EnumClass по заданным имени таблицы и поля.
func (e *EnumPostgres) GetEnumClasses(ctx context.Context, entity, field string) ([]entities.EnumClass, error) {
	if e.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectEnumClasses(entity, field)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build GetEnumClasses query: %w", err))
	}

	rows, err := e.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute GetEnumClasses query: %w", err))
	}
	defer rows.Close()

	enumClasses := make([]entities.EnumClass, 0)

	for rows.Next() {
		class, err := scanEnumClass(rows)
		if err != nil {
			return nil, err
		}
		enumClasses = append(enumClasses, *class)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("failed to fetch rows: %w", err))
	}

	return enumClasses, nil
}

// scanEnumClass сканирует строку результата запроса и преобразует ее в сущность EnumClass.
func scanEnumClass(row pgx.Row) (*entities.EnumClass, error) {
	var class entities.EnumClass
	var unit pgtype.Text

	err := row.Scan(&class.ID, &class.Type, &class.EntityName, &class.FieldName, &unit, &class.IsActive)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("scan EnumClass: %w", err))
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
	if e.Pool == nil {
		return 0, apperrors.Internal(errors.New("pool is nil"))
	}

	var valID uint

	enumValueRow, err := enumValue.ToRow()
	if err != nil || enumValueRow == nil {
		return 0, apperrors.Internal(fmt.Errorf("convert EnumValue to EnumValueRow: %w", err))
	}

	psql := queries.InsertEnumValue(*enumValueRow)

	query, args, err := psql.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("build InsertEnumValue query: %w", err))
	}

	err = e.Pool.QueryRow(ctx, query, args...).Scan(&valID)
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("execute InsertEnumValue query: %w", err))
	}

	if valID == 0 {
		return 0, apperrors.Internal(errors.New("failed to insert enum value"))
	}

	return valID, nil
}

// UpdateEnumValue обновляет сущность EnumValue в хранилище.
func (e *EnumPostgres) UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error {
	if e.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}

	if value, ok := updates["value_raw"]; ok {
		if value == nil {
			delete(updates, "value_raw")
		} else {
			valueStr := fmt.Sprintf("%v", value)
			updates["value_raw"] = valueStr
		}
	}

	psql := queries.UpdateEnumValue(id, updates)

	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("build UpdateEnumValue query: %w", err))
	}

	_, err = e.Pool.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("execute UpdateEnumValue query: %w", err))
	}

	return nil
}

// DeleteEnumValueByID удаляет сущность EnumValue из хранилища.
func (e *EnumPostgres) DeleteEnumValueByID(ctx context.Context, id uint) error {
	if e.Pool == nil {
		return apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.DeleteEnumValue(id)

	query, args, err := psql.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("build DeleteEnumValueByID query: %w", err))
	}

	_, err = e.Pool.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("execute DeleteEnumValueByID query: %w", err))
	}
	return nil
}

// GetEnumValues получает список сущностей EnumClass по заданным имени таблицы, поля и типу значения.
func (e *EnumPostgres) GetEnumValues(ctx context.Context, entity, field string, valueType entities.EnumType) ([]entities.EnumValue, error) {
	if e.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectEnumValues(entity, field, valueType)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build GetEnumValues query: %w", err))
	}

	rows, err := e.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute GetEnumValues query: %w", err))
	}
	defer rows.Close()

	enumValues := make([]entities.EnumValue, 0)

	for rows.Next() {
		val, err := scanEnumValue(rows)
		if err != nil {
			return nil, err
		}
		enumValues = append(enumValues, *val)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("failed to fetch rows: %w", err))
	}

	return enumValues, nil
}

// GetEnumValuesByClassID получает список сущностей EnumValue по заданному ID класса перечисления.
func (e *EnumPostgres) GetEnumValuesByClassID(ctx context.Context, classID uint) ([]entities.EnumValue, error) {
	if e.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectEnumValuesByClassID(classID)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build SelectEnumValuesByClassID query: %w", err))
	}

	rows, err := e.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute SelectEnumValuesByClassID query: %w", err))
	}
	defer rows.Close()

	enumValues := make([]entities.EnumValue, 0)

	for rows.Next() {
		val, err := scanEnumValue(rows)
		if err != nil {
			return nil, err
		}
		enumValues = append(enumValues, *val)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("failed to fetch rows: %w", err))
	}

	return enumValues, nil
}

// GetEnumClassByID получает сущность EnumClass по заданному ID класса перечисления.
func (e *EnumPostgres) GetEnumClassByID(ctx context.Context, id uint) (*entities.EnumClass, error) {
	if e.Pool == nil {
		return nil, apperrors.Internal(errors.New("pool is nil"))
	}

	psql := queries.SelectEnumClassByID(id)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build SelectEnumClassByID query: %w", err))
	}

	row := e.Pool.QueryRow(ctx, query, args...)
	class, err := scanEnumClass(row)
	if err != nil {
		return nil, err
	}

	return class, nil
}

// scanEnumValue сканирует строку результата запроса и преобразует ее в сущность EnumValue.
func scanEnumValue(row pgx.Row) (*entities.EnumValue, error) {
	var valRow entities.EnumValueRow

	err := row.Scan(&valRow.ID, &valRow.EnumClassID, &valRow.ValueRaw, &valRow.ValueType, &valRow.Position)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("scan EnumValue: %w", err))
	}

	val, err := valRow.FromRow()
	if err != nil || val == nil {
		return nil, apperrors.Internal(fmt.Errorf("convert EnumValueRow to EnumValue: %w", err))
	}

	return val, nil
}
