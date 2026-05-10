// Package repository содержит слой для манипуляции объектами в базе данных
package repository

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"context"
	"fmt"

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
	GetParameters(ctx context.Context, categoryID uint) ([]entities.NumericParameter, []entities.EnumParameter, error)

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
		return nil, err
	}

	var id uint
	err = p.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return nil, err
	}

	param.ID = id
	return param, nil
}

// UpdateNumericParameter обновляет существующий числовой параметр в базе данных и возвращает его.
func (p *ParameterPostgres) UpdateNumericParameter(ctx context.Context, id uint, updates map[string]any) (*entities.NumericParameter, error) {
	oldParam, err := p.fetchNumericParameterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	query := queries.UpdateNumericParameter(id, updates)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var param *entities.NumericParameter
	param, err = scanNumericParameter(p.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, err
	}

	if err := p.handleInheritableToggle(ctx, false, id, oldParam.Inheritable, param.Inheritable); err != nil {
		return nil, err
	}

	return param, nil
}

// DeleteNumericParameter удаляет числовой параметр из базы данных и возвращает его.
func (p *ParameterPostgres) DeleteNumericParameter(ctx context.Context, id uint) (*entities.NumericParameter, error) {
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

// InsertEnumParameter добавляет новый перечисляемый параметр в базу данных и возвращает его.
func (p *ParameterPostgres) InsertEnumParameter(ctx context.Context, param *entities.EnumParameter) (*entities.EnumParameter, error) {
	query := queries.InsertEnumParameter(param)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var id uint
	err = p.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return nil, err
	}

	param.ID = id
	return param, nil
}

// UpdateEnumParameter обновляет существующий перечисляемый параметр в базе данных и возвращает его.
func (p *ParameterPostgres) UpdateEnumParameter(ctx context.Context, id uint, updates map[string]any) (*entities.EnumParameter, error) {
	oldParam, err := p.fetchEnumParameterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	query := queries.UpdateEnumParameter(id, updates)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var param *entities.EnumParameter
	param, err = scanEnumParameter(p.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, err
	}

	if err := p.handleInheritableToggle(ctx, true, id, oldParam.Inheritable, param.Inheritable); err != nil {
		return nil, err
	}

	return param, nil
}

// DeleteEnumParameter удаляет перечисляемый параметр из базы данных и возвращает его.
func (p *ParameterPostgres) DeleteEnumParameter(ctx context.Context, id uint) (*entities.EnumParameter, error) {
	param, err := p.fetchEnumParameterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	deleteQuery := queries.DeleteEnumParameter(id)
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
func (p *ParameterPostgres) GetParameters(ctx context.Context, categoryID uint) ([]entities.NumericParameter, []entities.EnumParameter, error) {
	query := queries.SelectParameterIDsByCategory(categoryID, entities.MissingType)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, nil, err
	}

	var numericIDs []uint
	var enumIDs []uint
	err = p.Pool.QueryRow(ctx, sql, args...).Scan(&numericIDs, &enumIDs)
	if err != nil {
		return nil, nil, err
	}

	var numericParams []entities.NumericParameter
	var enumParams []entities.EnumParameter

	if len(numericIDs) > 0 {
		selectNumericQuery := queries.SelectNumericParameters(numericIDs)
		sql, args, err = selectNumericQuery.ToSql()
		if err != nil {
			return nil, nil, err
		}

		rows, err := p.Pool.Query(ctx, sql, args...)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var param *entities.NumericParameter
			param, err = scanNumericParameter(rows)
			if err != nil {
				return nil, nil, err
			}
			numericParams = append(numericParams, *param)
		}
	}

	if len(enumIDs) > 0 {
		selectEnumQuery := queries.SelectEnumParameters(enumIDs)
		sql, args, err = selectEnumQuery.ToSql()
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
func (p *ParameterPostgres) ApplyParameters(ctx context.Context, categoryID uint, numericParameters, enumParameters []int) (int, error) {
	rowsAffected := 0

	if len(numericParameters) > 0 {
		addQuery := queries.AddNumericParametersToCategories(numericParameters, []int{int(categoryID)})
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
		addQuery := queries.AddEnumParametersToCategories(enumParameters, []int{int(categoryID)})
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
	if err == nil {
		rowsAffected += inheritAffected
	}

	return rowsAffected, nil
}

// scanNumericParameter сканирует строку из базы данных в структуру NumericParameter и возвращает ее.
func scanNumericParameter(row pgx.Row) (*entities.NumericParameter, error) {
	var param entities.NumericParameter
	err := row.Scan(&param.ID, &param.MinValue,
		&param.MaxValue, &param.FieldName,
		&param.EntityName, &param.Inheritable)
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

// fetchNumericParameterByID получает NumericParameter по id
func (p *ParameterPostgres) fetchNumericParameterByID(ctx context.Context, id uint) (*entities.NumericParameter, error) {
	selectQuery := queries.SelectNumericParameters([]uint{id})
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}
	return scanNumericParameter(p.Pool.QueryRow(ctx, sql, args...))
}

// fetchEnumParameterByID получает EnumParameter по id
func (p *ParameterPostgres) fetchEnumParameterByID(ctx context.Context, id uint) (*entities.EnumParameter, error) {
	selectQuery := queries.SelectEnumParameters([]uint{id})
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}
	return scanEnumParameter(p.Pool.QueryRow(ctx, sql, args...))
}

// handleInheritableToggle реализует обработку изменения флага наследования у параметра
func (p *ParameterPostgres) handleInheritableToggle(ctx context.Context, isEnum bool, id uint, oldInheritable, newInheritable bool) error {
	if oldInheritable == newInheritable {
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

	if !oldInheritable && newInheritable {
		for _, categoryID := range categoryIDs {
			inheritQuery := queries.InheritParametersToChildren(categoryID)
			sql, args, err := inheritQuery.ToSql()
			if err != nil {
				return err
			}
			var affected int
			if err = p.Pool.QueryRow(ctx, sql, args...).Scan(&affected); err != nil {
				return err
			}
		}
		return nil
	}

	if isEnum {
		removeQuery := queries.RemoveEnumParametersFromDescendants([]int{int(id)}, categoryIDs)
		sql, args, err := removeQuery.ToSql()
		if err != nil {
			return err
		}
		var affected int
		if err = p.Pool.QueryRow(ctx, sql, args...).Scan(&affected); err != nil {
			return err
		}
		_ = affected
		return nil
	}

	removeQuery := queries.RemoveNumericParametersFromDescendants([]int{int(id)}, categoryIDs)
	sql, args, err = removeQuery.ToSql()
	if err != nil {
		return err
	}
	var affected int
	if err = p.Pool.QueryRow(ctx, sql, args...).Scan(&affected); err != nil {
		return err
	}
	_ = affected
	return nil
}
