// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"Brewery/pkg/logger"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ParameterPostgres реализует интерфейс ParameterRepository для работы с параметрами в базе данных PostgreSQL, используя пул соединений pgxpool.
type ParameterPostgres struct {
	Pool *pgxpool.Pool
}

// NewParameterPostgres создает новый экземпляр ParameterPostgres с предоставленным пулом соединений к базе данных.
func NewParameterPostgres(pool *pgxpool.Pool) *ParameterPostgres {
	return &ParameterPostgres{Pool: pool}
}

// InsertRangeParameter добавляет новый числовой параметр в базу данных и возвращает его.
func (p *ParameterPostgres) InsertRangeParameter(ctx context.Context, param *entities.RangeParameter) (*entities.RangeParameter, error) {
	query := queries.InsertNumericParameter(param)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var id int64
	err = p.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return nil, err
	}

	param.ID = id
	return param, nil
}

// UpdateRangeParameter обновляет существующий числовой параметр в базе данных и возвращает его.
func (p *ParameterPostgres) UpdateRangeParameter(ctx context.Context, id int64, updates map[string]any) (*entities.RangeParameter, error) {
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	oldParam, err := p.fetchNumericParameterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	query := queries.UpdateNumericParameter(id, updates)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var param *entities.RangeParameter
	param, err = scanNumericParameter(p.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, err
	}

	if err = p.toggleInheritance(ctx, false, id, oldParam.Inheritable, param.Inheritable); err != nil {
		return nil, err
	}

	return param, nil
}

// DeleteRangeParameter удаляет числовой параметр из базы данных и возвращает его.
func (p *ParameterPostgres) DeleteRangeParameter(ctx context.Context, id int64) (*entities.RangeParameter, error) {
	param, err := p.fetchNumericParameterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	deleteQuery := queries.DeleteNumericParameter(id)
	sql, args, err := deleteQuery.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return param, nil
}

// GetParameters извлекает все числовые и перечисляемые параметры из базы данных и возвращает их.
//
//nolint:funlen
func (p *ParameterPostgres) GetParameters(
	ctx context.Context,
	classID int64,
	parameterType int,
) (
	[]entities.RangeParameter,
	[]entities.EnumParameter,
	error,
) {
	var numericIDs []int64
	var enumIDs []int64
	var query sq.SelectBuilder

	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		return nil, nil, nil
	}

	if classID != 0 {
		query := queries.SelectParameterIDsByCategory(classID, parameterType)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, err
		}

		switch parameterType {
		case entities.MissingType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&numericIDs, &enumIDs)
		case entities.RangeParameterType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&numericIDs)
		case entities.EnumParameterType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&enumIDs)
		}
		if err != nil {
			return nil, nil, err
		}
	}

	var numericParams []entities.RangeParameter
	var enumParams []entities.EnumParameter

	if ((parameterType == entities.MissingType ||
		parameterType == entities.RangeParameterType) && classID == 0) || len(numericIDs) > 0 {
		log.Debug(ctx, "enum")
		query = queries.SelectNumericParameters(numericIDs)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, err
		}

		rows, err := p.Pool.Query(ctx, sql, args...)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var param *entities.RangeParameter
			param, err = scanNumericParameter(rows)
			if err != nil {
				return nil, nil, err
			}
			numericParams = append(numericParams, *param)
		}
	}
	if ((parameterType == entities.MissingType ||
		parameterType == entities.EnumParameterType) && classID == 0) || len(numericIDs) > 0 {
		query = queries.SelectEnumParameters(enumIDs)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, err
		}

		rows, err := p.Pool.Query(ctx, sql, args...)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var param *entities.EnumParameter
			param, err = scanEnumParameter(rows)
			if err != nil {
				return nil, nil, err
			}
			enumParams = append(enumParams, *param)
		}
	}

	return numericParams, enumParams, nil
}

// ApplyParameters применяет числовые и перечисляемые параметры к категории и возвращает результат применения.
func (p *ParameterPostgres) ApplyParameters(ctx context.Context, categoryID int64, numericParameters, enumParameters []int64) (int, error) {
	rowsAffected := 0

	if len(numericParameters) > 0 {
		addQuery := queries.AddNumericParametersToCategories(numericParameters, []int64{categoryID})
		sql, args, err := addQuery.ToSql()
		if err != nil {
			return 0, err
		}

		var affected int
		err = p.Pool.QueryRow(ctx, sql, args...).Scan(&affected)
		if err != nil {
			return 0, err
		}
		rowsAffected += affected
	}

	if len(enumParameters) > 0 {
		addQuery := queries.AddEnumParametersToCategories(enumParameters, []int64{categoryID})
		sql, args, err := addQuery.ToSql()
		if err != nil {
			return 0, err
		}

		var affected int
		err = p.Pool.QueryRow(ctx, sql, args...).Scan(&affected)
		if err != nil {
			return 0, err
		}
		rowsAffected += affected
	}

	inheritQuery := queries.InheritParametersToChildren(int(categoryID))
	sql, args, err := inheritQuery.ToSql()
	if err != nil {
		return 0, err
	}

	var inheritAffected int
	err = p.Pool.QueryRow(ctx, sql, args...).Scan(&inheritAffected)
	if err != nil {
		return 0, err
	}
	rowsAffected += inheritAffected

	return rowsAffected, nil
}

// scanNumericParameter сканирует строку из базы данных в структуру RangeParameter и возвращает ее.
func scanNumericParameter(row pgx.Row) (*entities.RangeParameter, error) {
	var param entities.RangeParameter
	err := row.Scan(&param.ID, &param.MinValue,
		&param.MaxValue, &param.FieldName, &param.Inheritable)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	return &param, nil
}

// scanEnumParameter сканирует строку из базы данных в структуру EnumParameter и возвращает ее.
func scanEnumParameter(row pgx.Row) (*entities.EnumParameter, error) {
	var param entities.EnumParameter
	err := row.Scan(&param.ID, &param.EnumClassID, &param.Inheritable)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Scan", err)
	}

	return &param, nil
}

// fetchNumericParameterByID получает RangeParameter по id
func (p *ParameterPostgres) fetchNumericParameterByID(ctx context.Context, id int64) (*entities.RangeParameter, error) {
	selectQuery := queries.SelectNumericParameters([]int64{id})
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}
	return scanNumericParameter(p.Pool.QueryRow(ctx, sql, args...))
}

// fetchEnumParameterByID получает EnumParameter по id
func (p *ParameterPostgres) fetchEnumParameterByID(ctx context.Context, id int64) (*entities.EnumParameter, error) {
	selectQuery := queries.SelectEnumParameters([]int64{id})
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}
	return scanEnumParameter(p.Pool.QueryRow(ctx, sql, args...))
}

// toggleInheritance реализует обработку изменения флага наследования у параметра
//
//nolint:funlen
func (p *ParameterPostgres) toggleInheritance(ctx context.Context, isEnum bool, id int64, oldValue, newValue bool) error {
	if oldValue == newValue {
		return nil
	}

	var categoriesQuery interface{ ToSql() (string, []any, error) }
	if isEnum {
		categoriesQuery = queries.SelectCategoriesWithEnumParameter(id)
	} else {
		categoriesQuery = queries.SelectCategoriesWithNumericParameter(id)
	}

	sql, args, err := categoriesQuery.ToSql()
	if err != nil {
		return err
	}

	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var categoryIDs []int
	for rows.Next() {
		var categoryID int
		if err = rows.Scan(&categoryID); err != nil {
			return err
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	if len(categoryIDs) == 0 {
		return nil
	}

	seen := make(map[int]struct{})
	uniq := make([]int, 0, len(categoryIDs))
	for _, cid := range categoryIDs {
		if _, ok := seen[cid]; !ok {
			seen[cid] = struct{}{}
			uniq = append(uniq, cid)
		}
	}

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if !oldValue && newValue {
		for _, categoryID := range uniq {
			inheritQuery := queries.InheritParametersToChildren(categoryID)
			sql, args, err = inheritQuery.ToSql()
			if err != nil {
				return err
			}

			if _, err = tx.Exec(ctx, sql, args...); err != nil {
				return err
			}
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
		return nil
	}

	if isEnum {
		removeQuery := queries.RemoveEnumParametersFromDescendants([]int{int(id)}, uniq)
		sql, args, err = removeQuery.ToSql()
		if err != nil {
			return err
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		return nil
	}

	removeQuery := queries.RemoveNumericParametersFromDescendants([]int{int(id)}, uniq)
	sql, args, err = removeQuery.ToSql()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *ParameterPostgres) ExcludeParameters(ctx context.Context, id int64, rangeParameters, enumParameters []int64) (int, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("Begin: %w", err)
	}

	defer tx.Rollback(ctx)

	qTx := queries.New(s.Pool).WithTx(tx)
	affected := 0

	if len(enumParameters) != 0 {
		params := queries.ExcludeEnumParametersParams{
			ProductID:        &id,
			EnumParameterIds: enumParameters,
		}
		if err := qTx.ExcludeEnumParameters(ctx, params); err != nil {
			return 0, fmt.Errorf("ExcludeEnumParameters: %w", err)
		}
		affected += len(enumParameters)
	}

	if len(rangeParameters) != 0 {
		params := queries.ExcludeRangeParametersParams{
			ProductID:         &id,
			RangeParameterIds: rangeParameters,
		}
		if err := qTx.ExcludeRangeParameters(ctx, params); err != nil {
			return 0, fmt.Errorf("ExcludeRangeParameters: %w", err)
		}
		affected += len(rangeParameters)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("Commit: %w", err)
	}

	return affected, nil
}
