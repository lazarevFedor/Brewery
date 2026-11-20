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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnumParametersPostgres хранит перечисляемые параметры в PostgreSQL.
type EnumParametersPostgres struct{ Pool *pgxpool.Pool }

// NewEnumParametersPostgres создает репозиторий перечисляемых параметров.
func NewEnumParametersPostgres(pool *pgxpool.Pool) *EnumParametersPostgres {
	return &EnumParametersPostgres{Pool: pool}
}

// InsertEnumParameter сохраняет новый перечисляемый параметр в хранилище.
func (e *EnumParametersPostgres) InsertEnumParameter(ctx context.Context, parameter entities.EnumClass) (uint, error) {
	if e.Pool == nil {
		return 0, errors.New("pool is nil")
	}

	var enumID uint

	psql := queries.InsertEnumParameter(parameter)

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

// UpdateEnumParameter обновляет перечисляемый параметр в хранилище.
func (e *EnumParametersPostgres) UpdateEnumParameter(ctx context.Context, id uint, updates map[string]any) error {
	if e.Pool == nil {
		return errors.New("pool is nil")
	}

	if len(updates) == 0 {
		return errors.New("no updates provided")
	}
	tx, err := e.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if lockErr := LockParameterDefinitions(ctx, tx); lockErr != nil {
		return lockErr
	}

	query, args, err := queries.SelectEnumParameterByID(id).Suffix("FOR UPDATE").ToSql()
	if err != nil {
		return err
	}
	oldParameter, err := scanEnumParameter(tx.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.NotFound("Enum parameter not found", err)
	}
	if err != nil {
		return err
	}

	query, args, err = queries.UpdateEnumParameter(id, updates).Suffix("RETURNING inheritable").ToSql()
	if err != nil {
		return err
	}
	var inheritable bool
	if err := tx.QueryRow(ctx, query, args...).Scan(&inheritable); err != nil {
		return ParameterFieldNameError(err)
	}
	if err := toggleParameterInheritance(ctx, tx, true, int64(id), oldParameter.Inheritable, inheritable); err != nil {
		return ParameterFieldNameError(err)
	}
	return ParameterFieldNameError(tx.Commit(ctx))
}

// DeleteEnumParameterByID удаляет перечисляемый параметр из хранилища.
func (e *EnumParametersPostgres) DeleteEnumParameterByID(ctx context.Context, id uint) error {
	if e.Pool == nil {
		return errors.New("pool is nil")
	}

	psql := queries.DeleteEnumParameter(id)

	query, args, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", "ToSql", err)
	}

	result, err := e.Pool.Exec(ctx, query, args...)
	if err != nil {
		return parameterDeleteError(err, "Enum parameter")
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("Enum parameter not found", nil)
	}

	return nil
}

func parameterDeleteError(err error, resource string) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && (pgErr.Code == "23503" || pgErr.Code == "23001") {
		return apperrors.Conflict(resource+" is referenced by another record", err)
	}
	return fmt.Errorf("delete %s: %w", resource, err)
}

// GetEnumParameters получает перечисляемые параметры по имени поля.
func (e *EnumParametersPostgres) GetEnumParameters(ctx context.Context, field string) ([]entities.EnumClass, error) {
	if e.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectEnumParameters(field)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	rows, err := e.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Query", err)
	}
	defer rows.Close()

	parameters := make([]entities.EnumClass, 0)

	for rows.Next() {
		parameter, err := scanEnumParameter(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		parameters = append(parameters, *parameter)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return parameters, nil
}

// scanEnumParameter сканирует строку результата запроса и преобразует ее в перечисляемый параметр.
func scanEnumParameter(row pgx.Row) (*entities.EnumClass, error) {
	var parameter entities.EnumClass
	var unit pgtype.Text

	err := row.Scan(&parameter.ID, &parameter.Type, &parameter.FieldName, &unit, &parameter.Inheritable)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	if unit.Valid {
		parameter.Unit = unit.String
	} else {
		parameter.Unit = ""
	}

	return &parameter, nil
}

// GetEnumParameterByID получает перечисляемый параметр по ID.
func (e *EnumParametersPostgres) GetEnumParameterByID(ctx context.Context, id uint) (*entities.EnumClass, error) {
	if e.Pool == nil {
		return nil, errors.New("pool is nil")
	}

	psql := queries.SelectEnumParameterByID(id)

	query, args, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToSql", err)
	}

	row := e.Pool.QueryRow(ctx, query, args...)
	parameter, err := scanEnumParameter(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "ToRow", err)
	}

	return parameter, nil
}
