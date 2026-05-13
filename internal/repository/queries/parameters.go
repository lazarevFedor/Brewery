// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

const (
	NumericParametersTable = "numeric_parameters"
	EnumParametersTable    = "enum_parameters"

	NumericReturningFields = "id, min_val, max_val, field_name, entity_name, inheritable"
	EnumReturningFields    = "id, enum_class_id, inheritable"
)

// InsertNumericParameter возвращает запрос для вставки
// нового числового параметра в базу данных.
func InsertNumericParameter(param *entities.NumericParameter) sq.InsertBuilder {
	data := map[string]any{
		"min_val":     param.MinValue,
		"max_val":     param.MaxValue,
		"field_name":  param.FieldName,
		"entity_name": param.EntityName,
		"inheritable": param.Inheritable,
	}

	return psql.
		Insert(NumericParametersTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// UpdateNumericParameter возвращает запрос для обновления
// существующего числового параметра в базе данных.
func UpdateNumericParameter(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(NumericParametersTable).
		SetMap(updates).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + NumericReturningFields)
}

// DeleteNumericParameter возвращает запрос для
// удаления числового параметра из базы данных.
func DeleteNumericParameter(id uint) sq.DeleteBuilder {
	return psql.
		Delete(NumericParametersTable).
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
func SelectParameterIDsByCategory(categoryID uint, parameterType int) sq.SelectBuilder {
	var query sq.SelectBuilder
	switch parameterType {
	case entities.NumericParameterType:
		query = psql.Select("numeric_parameter_ids")

	case entities.EnumParameterType:
		query = psql.Select("enum_parameter_ids")

	default:
		query = psql.Select("numeric_parameter_ids, enum_parameter_ids")
	}

	query = query.From(tableCategories)
	if categoryID != 0 {
		query = query.Where(sq.Eq{"id": categoryID})
	}

	return query
}

// SelectNumericParameters возвращает запрос для получения всех параметров, без фильтрации по категории,
// включая как числовые параметры.
func SelectNumericParameters(ids []uint) sq.SelectBuilder {
	if len(ids) == 0 {
		return psql.
			Select(NumericReturningFields).
			From(NumericParametersTable)
	}

	return psql.
		Select(NumericReturningFields).
		From(NumericParametersTable).
		Where(sq.Eq{"id": ids})
}

// SelectEnumParameters возвращает запрос для получения всех параметров, без фильтрации по категории
func SelectEnumParameters(ids []uint) sq.SelectBuilder {
	if len(ids) == 0 {
		return psql.
			Select(EnumReturningFields).
			From(EnumParametersTable)
	}

	return psql.
		Select(EnumReturningFields).
		From(EnumParametersTable).
		Where(sq.Eq{"id": ids})
}

// AddNumericParametersToCategories возвращает запрос для добавления числовых параметров в категории
func AddNumericParametersToCategories(parameterIDs []int, categoryIDs []int) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"add_numeric_parameters_to_categories(?::int[], ?::int[])",
				pq.Array(parameterIDs),
				pq.Array(categoryIDs),
			),
		)
}

// AddEnumParametersToCategories возвращает запрос для добавления параметров-перечислений в категории
func AddEnumParametersToCategories(parameterIDs []int, categoryIDs []int) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"add_enum_parameters_to_categories(?::int[], ?::int[])",
				pq.Array(parameterIDs),
				pq.Array(categoryIDs),
			),
		)
}

// InheritParametersToChildren возвращает запрос на наследование параметров дочерними категориями
// Рекурсивно находит всех потомков и добавляет им наследуемые параметры
func InheritParametersToChildren(parentID int) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"inherit_parameters_to_children(?)",
				parentID,
			),
		)
}

// SelectCategoriesWithNumericParameter возвращает запрос для получения всех категорий, содержащих числовой параметр
func SelectCategoriesWithNumericParameter(parameterID uint) sq.SelectBuilder {
	return psql.
		Select("id").
		From(tableCategories).
		Where(sq.Expr("numeric_parameter_ids @> ARRAY[?::INT]", int(parameterID)))
}

// SelectCategoriesWithEnumParameter возвращает запрос для получения всех категорий, содержащих параметр-перечисление
func SelectCategoriesWithEnumParameter(parameterID uint) sq.SelectBuilder {
	return psql.
		Select("id").
		From(tableCategories).
		Where(sq.Expr("enum_parameter_ids @> ARRAY[?::INT]", int(parameterID)))
}

// RemoveNumericParametersFromDescendants возвращает запрос для удаления числовых параметров из потомков категорий
func RemoveNumericParametersFromDescendants(parameterIDs []int, parentIDs []int) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"remove_numeric_parameters_from_descendants(?::int[], ?::int[])",
				pq.Array(parameterIDs),
				pq.Array(parentIDs),
			),
		)
}

// RemoveEnumParametersFromDescendants возвращает запрос для удаления параметров-перечислений из потомков категорий
func RemoveEnumParametersFromDescendants(parameterIDs []int, parentIDs []int) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"remove_enum_parameters_from_descendants(?::int[], ?::int[])",
				pq.Array(parameterIDs),
				pq.Array(parentIDs),
			),
		)
}
