// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnumValuesPostgres хранит значения перечисляемых параметров в PostgreSQL.
type EnumValuesPostgres struct{ Pool *pgxpool.Pool }

// NewEnumValuesPostgres создает репозиторий значений перечислений.
func NewEnumValuesPostgres(pool *pgxpool.Pool) *EnumValuesPostgres {
	return &EnumValuesPostgres{Pool: pool}
}

// InsertEnumValue сохраняет новую сущность EnumValue в хранилище.
func (e *EnumValuesPostgres) InsertEnumValue(ctx context.Context, enumValue entities.EnumValue) (uint, error) {
	if e.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	tx, err := e.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var valueType entities.EnumType
	err = tx.QueryRow(ctx, "SELECT enum_type FROM enum_parameters WHERE id = $1 FOR NO KEY UPDATE", enumValue.EnumClassID).Scan(&valueType)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, apperrors.BadRequest("enum parameter not found", err)
	}
	if err != nil {
		return 0, err
	}
	if enumValue.ValueType != valueType {
		return 0, apperrors.BadRequest("enum_type must match the parameter definition", nil)
	}

	enumValueRow, err := enumValue.ToRow()
	if err != nil {
		return 0, apperrors.BadRequest("invalid enum value for the parameter type", err)
	}

	psql := queries.InsertEnumValue(*enumValueRow)

	query, args, err := psql.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "ToSql", err)
	}

	var valID uint
	err = tx.QueryRow(ctx, query, args...).Scan(&valID)
	if err != nil {
		return 0, enumValueWriteError(err)
	}

	if valID == 0 {
		return 0, errors.New("failed to insert enum value")
	}
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return 0, commitErr
	}

	return valID, nil
}

// UpdateEnumValue обновляет сущность EnumValue в хранилище.
func (e *EnumValuesPostgres) UpdateEnumValue(ctx context.Context, id uint, updates map[string]any) error {
	if e.Pool == nil {
		return errors.New("pool is nil")
	}

	tx, err := e.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	valueType, err := lockEnumValueType(ctx, tx, id)
	if err != nil {
		return err
	}
	if value, ok := updates["value_raw"]; ok {
		enumValue := entities.EnumValue{Value: value, ValueType: valueType}
		row, encodeErr := enumValue.ToRow()
		if encodeErr != nil {
			return apperrors.BadRequest("invalid enum value for the parameter type", encodeErr)
		}
		updates = maps.Clone(updates)
		updates["value_raw"] = row.ValueRaw
	}

	psql := queries.UpdateEnumValue(id, updates)

	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return enumValueWriteError(err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("Enum value not found", nil)
	}

	return tx.Commit(ctx)
}

// DeleteEnumValueByID удаляет сущность EnumValue из хранилища.
func (e *EnumValuesPostgres) DeleteEnumValueByID(ctx context.Context, id uint) error {
	if e.Pool == nil {
		return errors.New("pool is nil")
	}

	tx, err := e.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, lockErr := lockEnumValueType(ctx, tx, id); lockErr != nil {
		return lockErr
	}

	psql := queries.DeleteEnumValue(id)
	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return parameterDeleteError(err, "Enum value")
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("Enum value not found", nil)
	}
	return tx.Commit(ctx)
}

// Keep the definition stable during the write and serialize changes to the value order.
func lockEnumValueType(ctx context.Context, tx pgx.Tx, id uint) (entities.EnumType, error) {
	const query = `SELECT p.enum_type
FROM enum_parameters p JOIN parameter_values v ON v.enum_parameter_id = p.id
WHERE v.id = $1
FOR NO KEY UPDATE OF p`
	var valueType entities.EnumType
	err := tx.QueryRow(ctx, query, id).Scan(&valueType)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", apperrors.NotFound("Enum value not found", err)
	}
	return valueType, err
}

func enumValueWriteError(err error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch {
		case pgErr.Code == "23503" && pgErr.ConstraintName == "parameter_values_enum_parameter_id_fkey":
			return apperrors.BadRequest("enum parameter not found", err)
		case pgErr.Code == "22001":
			return apperrors.BadRequest("enum value exceeds the allowed length", err)
		}
	}
	return fmt.Errorf("write enum value: %w", err)
}

// GetEnumValues получает значения перечисления по имени поля и типу значения.
func (e *EnumValuesPostgres) GetEnumValues(ctx context.Context, field string, valueType entities.EnumType) ([]entities.EnumValue, error) {
	if e.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectEnumValues(field, valueType)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	rows, err := e.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Query", err)
	}
	defer rows.Close()

	enumValues := make([]entities.EnumValue, 0)

	for rows.Next() {
		val, err := scanEnumValue(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		enumValues = append(enumValues, *val)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return enumValues, nil
}

// GetEnumValuesByParameterID получает список сущностей EnumValue по ID перечисляемого параметра.
func (e *EnumValuesPostgres) GetEnumValuesByParameterID(ctx context.Context, parameterID uint) ([]entities.EnumValue, error) {
	if e.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectEnumValuesByParameterID(parameterID)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	rows, err := e.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Query", err)
	}
	defer rows.Close()

	enumValues := make([]entities.EnumValue, 0)

	for rows.Next() {
		val, err := scanEnumValue(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		enumValues = append(enumValues, *val)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return enumValues, nil
}

// scanEnumValue сканирует строку результата запроса и преобразует ее в сущность EnumValue.
func scanEnumValue(row pgx.Row) (*entities.EnumValue, error) {
	var valRow entities.ParameterValueRow

	err := row.Scan(&valRow.ID, &valRow.EnumClassID, &valRow.ValueRaw, &valRow.ValueType, &valRow.Position)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	val, err := valRow.FromRow()
	if err != nil || val == nil {
		return nil, fmt.Errorf("%s: %w", "FromRow", err)
	}

	return val, nil
}
