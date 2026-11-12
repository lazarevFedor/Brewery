// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

const (
	RangeParametersTable = "range_parameters"
	EnumParametersTable  = "enum_parameters"

	NumericReturningFields = "id, min_val, max_val, field_name, inheritable, unit"
	EnumReturningFields    = "id, id AS enum_class_id, inheritable"
)

// InsertNumericParameter возвращает запрос для вставки
// нового числового параметра в базу данных.
func InsertNumericParameter(param *entities.RangeParameter) sq.InsertBuilder {
	data := map[string]any{
		"min_val":     param.MinValue,
		"max_val":     param.MaxValue,
		"field_name":  param.FieldName,
		"inheritable": param.Inheritable,
		"unit":        param.Unit,
	}

	return psql.
		Insert(RangeParametersTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// UpdateNumericParameter возвращает запрос для обновления
// существующего числового параметра в базе данных.
func UpdateNumericParameter(id int64, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(RangeParametersTable).
		SetMap(updates).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + NumericReturningFields)
}

// DeleteNumericParameter возвращает запрос для
// удаления числового параметра из базы данных.
func DeleteNumericParameter(id int64) sq.DeleteBuilder {
	return psql.
		Delete(RangeParametersTable).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + NumericReturningFields)
}

// InsertEnumParameter возвращает запрос для вставки
// нового параметра-перечисления в базу данных.
func InsertEnumParameter(param *entities.EnumParameter) sq.InsertBuilder {
	data := map[string]any{
		"enum_class_id": param.EnumClassID,
		"inheritable":   param.Inheritable,
	}

	return psql.
		Insert(EnumParametersTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// UpdateEnumParameter возвращает запрос для обновления
// существующего параметра-перечисления в базе данных.
func UpdateEnumParameter(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(EnumParametersTable).
		SetMap(updates).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + EnumReturningFields)
}

// DeleteEnumParameter возвращает запрос для
// удаления параметра-перечисления из базы данных.
func DeleteEnumParameter(id uint) sq.DeleteBuilder {
	return psql.
		Delete(EnumParametersTable).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + EnumReturningFields)
}

// SelectParameterIDsByCategory возвращает запрос для получения всех параметров,
// связанных с определенной категорией, включая как числовые, так и перечисляемые параметры.
func SelectParameterIDsByCategory(categoryID int64, parameterType int) sq.SelectBuilder {
	var query sq.SelectBuilder
	switch parameterType {
	case entities.RangeParameterType:
		query = psql.Select("numeric_parameter_ids")

	case entities.EnumParameterType:
		query = psql.Select("enum_parameter_ids")

	case entities.MissingType:
		query = psql.Select("numeric_parameter_ids, enum_parameter_ids")
	}

	query = query.From(tableClasses)
	if categoryID != 0 {
		query = query.Where(sq.Eq{"id": categoryID})
	}

	return query
}

// SelectNumericParameters возвращает запрос для получения всех параметров, без фильтрации по категории,
// включая как числовые параметры.
func SelectNumericParameters(ids []int64) sq.SelectBuilder {
	query := psql.
		Select(NumericReturningFields).
		From(RangeParametersTable)

	if len(ids) > 0 {
		query = query.Where(sq.Eq{"id": ids})
	}
	return query
}

// SelectEnumParameters возвращает запрос для получения всех параметров, без фильтрации по категории
func SelectEnumParameters(ids []int64) sq.SelectBuilder {
	query := psql.
		Select(EnumReturningFields).
		From(EnumParametersTable)

	if len(ids) > 0 {
		query = query.Where(sq.Eq{"id": ids})
	}
	return query
}

// AddNumericParametersToCategories возвращает запрос для добавления числовых параметров в категории
func AddNumericParametersToCategories(parameterIDs []int64, categoryIDs []int64) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"add_numeric_parameters_to_categories(?::bigint[], ?::bigint[])",
				pq.Array(parameterIDs),
				pq.Array(categoryIDs),
			),
		)
}

// AddEnumParametersToCategories возвращает запрос для добавления параметров-перечислений в категории
func AddEnumParametersToCategories(parameterIDs []int64, categoryIDs []int64) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"add_enum_parameters_to_categories(?::bigint[], ?::bigint[])",
				pq.Array(parameterIDs),
				pq.Array(categoryIDs),
			),
		)
}

// InheritParametersToChildren возвращает запрос на наследование параметров дочерними категориями
// Рекурсивно находит всех потомков и добавляет им наследуемые параметры
func InheritParametersToChildren(parentID int64) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"inherit_parameters_to_children(?::bigint)",
				parentID,
			),
		)
}

// SelectCategoriesWithNumericParameter возвращает запрос для получения всех категорий, содержащих числовой параметр
func SelectCategoriesWithNumericParameter(parameterID int64) sq.SelectBuilder {
	return psql.
		Select("id").
		From(tableClasses).
		Where(sq.Expr("numeric_parameter_ids @> ARRAY[?::bigint]", parameterID))
}

// SelectCategoriesWithEnumParameter возвращает запрос для получения всех категорий, содержащих параметр-перечисление
func SelectCategoriesWithEnumParameter(parameterID int64) sq.SelectBuilder {
	return psql.
		Select("id").
		From(tableClasses).
		Where(sq.Expr("enum_parameter_ids @> ARRAY[?::bigint]", parameterID))
}

// RemoveNumericParametersFromDescendants возвращает запрос для удаления числовых параметров из потомков категорий
func RemoveNumericParametersFromDescendants(parameterIDs []int64, parentIDs []int64) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"remove_numeric_parameters_from_descendants(?::bigint[], ?::bigint[])",
				pq.Array(parameterIDs),
				pq.Array(parentIDs),
			),
		)
}

// RemoveEnumParametersFromDescendants возвращает запрос для удаления параметров-перечислений из потомков категорий
func RemoveEnumParametersFromDescendants(parameterIDs []int64, parentIDs []int64) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"remove_enum_parameters_from_descendants(?::bigint[], ?::bigint[])",
				pq.Array(parameterIDs),
				pq.Array(parentIDs),
			),
		)
}
