// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/apperrors"
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

// ParameterRepository определяет интерфейс для работы с параметрами, включая числовые и перечисляемые параметры.
type ParameterRepository interface {
	// InsertNumericParameter добавляет новый числовой параметр в базу данных и возвращает его.
	InsertNumericParameter(ctx context.Context, param *entities.NumericParameter) (*entities.NumericParameter, error)

	// UpdateNumericParameter обновляет существующий числовой параметр в базе данных и возвращает его.
	UpdateNumericParameter(ctx context.Context, id uint, updates map[string]any) (*entities.NumericParameter, error)

	// DeleteNumericParameter удаляет числовой параметр из базы данных и возвращает его.
	DeleteNumericParameter(ctx context.Context, id uint) (*entities.NumericParameter, error)

	// InsertEnumParameter добавляет новый перечисляемый параметр в базу данных и возвращает его.
	InsertEnumParameter(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error)

	// UpdateEnumParameter обновляет существующий перечисляемый параметр в базе данных и возвращает его.
	UpdateEnumParameter(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error)

	// DeleteEnumParameter удаляет перечисляемый параметр из базы данных и возвращает его.
	DeleteEnumParameter(ctx context.Context, id uint) (*entities.EnumParameter, error)

	// GetParameters извлекает все числовые и перечисляемые параметры из базы данных и возвращает их.
	GetParameters(ctx context.Context, categoryID uint, parameterType int) ([]entities.NumericParameter, []entities.EnumParameter, error)

	// ApplyParameters применяет числовые и перечисляемые параметры к категории и возвращает результат применения.
	ApplyParameters(ctx context.Context, categoryID uint, numericParameters, enumParameters []int) (int, error)
}

// ParameterPostgres реализует интерфейс ParameterRepository для работы с параметрами в базе данных PostgreSQL, используя пул соединений pgxpool.
type ParameterPostgres struct {
	Pool *pgxpool.Pool
}

// NewParameterRepository создает новый экземпляр ParameterRepository с предоставленным пулом соединений к базе данных.
func NewParameterRepository(pool *pgxpool.Pool) ParameterRepository {
	return &ParameterPostgres{Pool: pool}
}

// NewParameterPostgres создает новый экземпляр ParameterPostgres с предоставленным пулом соединений к базе данных.
func NewParameterPostgres(pool *pgxpool.Pool) *ParameterPostgres {
	return &ParameterPostgres{Pool: pool}
}

// InsertNumericParameter добавляет новый числовой параметр в базу данных и возвращает его.
func (p *ParameterPostgres) InsertNumericParameter(ctx context.Context, param *entities.NumericParameter) (*entities.NumericParameter, error) {
	query := queries.InsertNumericParameter(param)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build insert numeric parameter query: %w", err))
	}

	var id uint
	err = p.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute insert numeric parameter query: %w", err))
	}

	param.ID = id
	return param, nil
}

// UpdateNumericParameter обновляет существующий числовой параметр в базе данных и возвращает его.
func (p *ParameterPostgres) UpdateNumericParameter(ctx context.Context, id uint, updates map[string]any) (*entities.NumericParameter, error) {
	if len(updates) == 0 {
		return nil, apperrors.BadRequest("no updates provided", errors.New("no updates provided"))
	}

	oldParam, err := p.fetchNumericParameterByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("fetch numeric parameter by id: %w", err))
	}

	query := queries.UpdateNumericParameter(id, updates)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build update numeric parameter query: %w", err))
	}

	var param *entities.NumericParameter
	param, err = scanNumericParameter(p.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute update numeric parameter query: %w", err))
	}

	if err = p.toggleInheritance(ctx, false, id, oldParam.Inheritable, param.Inheritable); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toggle inheritance: %w", err))
	}

	return param, nil
}

// DeleteNumericParameter удаляет числовой параметр из базы данных и возвращает его.
func (p *ParameterPostgres) DeleteNumericParameter(ctx context.Context, id uint) (*entities.NumericParameter, error) {
	param, err := p.fetchNumericParameterByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("fetch numeric parameter by id: %w", err))
	}

	deleteQuery := queries.DeleteNumericParameter(id)
	sql, args, err := deleteQuery.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build delete numeric parameter query: %w", err))
	}

	_, err = p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute delete numeric parameter query: %w", err))
	}

	return param, nil
}

// InsertEnumParameter добавляет новый перечисляемый параметр в базу данных и возвращает его.
func (p *ParameterPostgres) InsertEnumParameter(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error) {
	query := queries.InsertEnumParameter(param)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build insert enum parameter query: %w", err))
	}

	var id uint
	err = p.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute insert enum parameter query: %w", err))
	}

	param.ID = id
	return param, nil
}

// UpdateEnumParameter обновляет существующий перечисляемый параметр в базе данных и возвращает его.
func (p *ParameterPostgres) UpdateEnumParameter(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error) {
	if len(updates) == 0 {
		return nil, apperrors.BadRequest("no updates provided", errors.New("no updates provided"))
	}

	oldParam, err := p.fetchEnumParameterByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("fetch enum parameter by id: %w", err))
	}

	query := queries.UpdateEnumParameter(id, updates)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build update enum parameter query: %w", err))
	}

	var param *entities.EnumParameter
	param, err = scanEnumParameter(p.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute update enum parameter query: %w", err))
	}

	if err = p.toggleInheritance(ctx, true, id, oldParam.Inheritable, param.Inheritable); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("toggle inheritance: %w", err))
	}

	return param, nil
}

// DeleteEnumParameter удаляет перечисляемый параметр из базы данных и возвращает его.
func (p *ParameterPostgres) DeleteEnumParameter(ctx context.Context, id uint) (*entities.EnumParameter, error) {
	param, err := p.fetchEnumParameterByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("fetch enum parameter by id: %w", err))
	}

	deleteQuery := queries.DeleteEnumParameter(id)
	sql, args, err := deleteQuery.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build delete enum parameter query: %w", err))
	}

	_, err = p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("execute delete enum parameter query: %w", err))
	}

	return param, nil
}

// GetParameters извлекает все числовые и перечисляемые параметры из базы данных и возвращает их.
//
//nolint:funlen
func (p *ParameterPostgres) GetParameters(
	ctx context.Context,
	categoryID uint,
	parameterType int,
) (
	[]entities.NumericParameter,
	[]entities.EnumParameter,
	error,
) {
	var numericIDs []uint
	var enumIDs []uint
	var query sq.SelectBuilder

	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		return nil, nil, nil
	}

	if categoryID != 0 {
		query = queries.SelectParameterIDsByCategory(categoryID, parameterType)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, apperrors.Internal(fmt.Errorf("build select parameter ids by category query: %w", err))
		}

		switch parameterType {
		case entities.MissingType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&numericIDs, &enumIDs)
		case entities.NumericParameterType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&numericIDs)
		case entities.EnumParameterType:
			err = p.Pool.QueryRow(ctx, sql, args...).Scan(&enumIDs)
		}
		if err != nil {
			return nil, nil, apperrors.Internal(fmt.Errorf("execute select parameter ids by category query: %w", err))
		}
	}

	var numericParams []entities.NumericParameter
	var enumParams []entities.EnumParameter

	if ((parameterType == entities.MissingType ||
		parameterType == entities.NumericParameterType) && categoryID == 0) || len(numericIDs) > 0 {
		log.Debug(ctx, "enum")
		query = queries.SelectNumericParameters(numericIDs)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, apperrors.Internal(fmt.Errorf("build select numeric parameters query: %w", err))
		}

		rows, err := p.Pool.Query(ctx, sql, args...)
		if err != nil {
			return nil, nil, apperrors.Internal(fmt.Errorf("execute select numeric parameters query: %w", err))
		}
		defer rows.Close()

		for rows.Next() {
			var param *entities.NumericParameter
			param, err = scanNumericParameter(rows)
			if err != nil {
				return nil, nil, apperrors.Internal(fmt.Errorf("scan numeric parameter: %w", err))
			}
			numericParams = append(numericParams, *param)
		}
	}
	if ((parameterType == entities.MissingType ||
		parameterType == entities.EnumParameterType) && categoryID == 0) || len(numericIDs) > 0 {
		query = queries.SelectEnumParameters(enumIDs)
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil, apperrors.Internal(fmt.Errorf("build select enum parameters query: %w", err))
		}

		rows, err := p.Pool.Query(ctx, sql, args...)
		if err != nil {
			return nil, nil, apperrors.Internal(fmt.Errorf("execute select enum parameters query: %w", err))
		}
		defer rows.Close()

		for rows.Next() {
			var param *entities.EnumParameter
			param, err = scanEnumParameter(rows)
			if err != nil {
				return nil, nil, apperrors.Internal(fmt.Errorf("scan enum parameter: %w", err))
			}
			enumParams = append(enumParams, *param)
		}
	}

	return numericParams, enumParams, nil
}

// ApplyParameters применяет числовые и перечисляемые параметры к категории и возвращает результат применения.
func (p *ParameterPostgres) ApplyParameters(ctx context.Context, categoryID uint, numericParameters, enumParameters []int) (int, error) {
	rowsAffected := 0

	if len(numericParameters) > 0 {
		addQuery := queries.AddNumericParametersToCategories(numericParameters, []int{int(categoryID)})
		sql, args, err := addQuery.ToSql()
		if err != nil {
			return 0, apperrors.Internal(fmt.Errorf("build AddNumericParametersToCategories query: %w", err))
		}

		var affected int
		err = p.Pool.QueryRow(ctx, sql, args...).Scan(&affected)
		if err != nil {
			return 0, apperrors.Internal(fmt.Errorf("execute AddNumericParametersToCategories query: %w", err))
		}
		rowsAffected += affected
	}

	if len(enumParameters) > 0 {
		addQuery := queries.AddEnumParametersToCategories(enumParameters, []int{int(categoryID)})
		sql, args, err := addQuery.ToSql()
		if err != nil {
			return 0, apperrors.Internal(fmt.Errorf("build AddEnumParametersToCategories query: %w", err))
		}

		var affected int
		err = p.Pool.QueryRow(ctx, sql, args...).Scan(&affected)
		if err != nil {
			return 0, apperrors.Internal(fmt.Errorf("execute AddEnumParametersToCategories query: %w", err))
		}
		rowsAffected += affected
	}

	inheritQuery := queries.InheritParametersToChildren(int(categoryID))
	sql, args, err := inheritQuery.ToSql()
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("build InheritParametersToChildren query: %w", err))
	}

	var inheritAffected int
	err = p.Pool.QueryRow(ctx, sql, args...).Scan(&inheritAffected)
	if err != nil {
		return 0, apperrors.Internal(fmt.Errorf("execute InheritParametersToChildren query: %w", err))
	}
	rowsAffected += inheritAffected

	return rowsAffected, nil
}

// scanNumericParameter сканирует строку из базы данных в структуру NumericParameter и возвращает ее.
func scanNumericParameter(row pgx.Row) (*entities.NumericParameter, error) {
	var param entities.NumericParameter
	err := row.Scan(&param.ID, &param.MinValue,
		&param.MaxValue, &param.FieldName,
		&param.EntityName, &param.Inheritable)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("scan numeric parameter: %w", err))
	}

	return &param, nil
}

// scanEnumParameter сканирует строку из базы данных в структуру EnumParameter и возвращает ее.
func scanEnumParameter(row pgx.Row) (*entities.EnumParameter, error) {
	var param entities.EnumParameter
	err := row.Scan(&param.ID, &param.EnumClassID, &param.Inheritable)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("scan enum parameter: %w", err))
	}

	return &param, nil
}

// fetchNumericParameterByID получает NumericParameter по id
func (p *ParameterPostgres) fetchNumericParameterByID(ctx context.Context, id uint) (*entities.NumericParameter, error) {
	selectQuery := queries.SelectNumericParameters([]uint{id})
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build select numeric parameter by id query: %w", err))
	}
	return scanNumericParameter(p.Pool.QueryRow(ctx, sql, args...))
}

// fetchEnumParameterByID получает EnumParameter по id
func (p *ParameterPostgres) fetchEnumParameterByID(ctx context.Context, id uint) (*entities.EnumParameter, error) {
	selectQuery := queries.SelectEnumParameters([]uint{id})
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("build select enum parameter by id query: %w", err))
	}
	return scanEnumParameter(p.Pool.QueryRow(ctx, sql, args...))
}

// toggleInheritance реализует обработку изменения флага наследования у параметра
//
//nolint:funlen
func (p *ParameterPostgres) toggleInheritance(ctx context.Context, isEnum bool, id uint, oldValue, newValue bool) error {
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
		return apperrors.Internal(fmt.Errorf("build select categories with parameter query: %w", err))
	}

	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("execute select categories with parameter query: %w", err))
	}
	defer rows.Close()

	var categoryIDs []int
	for rows.Next() {
		var categoryID int
		if err = rows.Scan(&categoryID); err != nil {
			return apperrors.Internal(fmt.Errorf("scan category id: %w", err))
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
		return apperrors.Internal(fmt.Errorf("begin transaction: %w", err))
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if !oldValue && newValue {
		for _, categoryID := range uniq {
			inheritQuery := queries.InheritParametersToChildren(categoryID)
			sql, args, err = inheritQuery.ToSql()
			if err != nil {
				return apperrors.Internal(fmt.Errorf("build InheritParametersToChildren query: %w", err))
			}

			if _, err = tx.Exec(ctx, sql, args...); err != nil {
				return apperrors.Internal(fmt.Errorf("execute InheritParametersToChildren query: %w", err))
			}
		}
		if err = tx.Commit(ctx); err != nil {
			return apperrors.Internal(fmt.Errorf("commit transaction: %w", err))
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
			return apperrors.Internal(fmt.Errorf("execute RemoveEnumParametersFromDescendants query: %w", err))
		}

		if err = tx.Commit(ctx); err != nil {
			return apperrors.Internal(fmt.Errorf("commit transaction: %w", err))
		}

		return nil
	}

	removeQuery := queries.RemoveNumericParametersFromDescendants([]int{int(id)}, uniq)
	sql, args, err = removeQuery.ToSql()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("build RemoveNumericParametersFromDescendants query: %w", err))
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return apperrors.Internal(fmt.Errorf("execute RemoveNumericParametersFromDescendants query: %w", err))
	}

	if err = tx.Commit(ctx); err != nil {
		return apperrors.Internal(fmt.Errorf("commit transaction: %w", err))
	}

	return nil
}
