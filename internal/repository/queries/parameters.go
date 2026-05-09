// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
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
		"id":          param.ID,
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
		Where(sq.Eq{"id": id})
}

// InsertEnumParameter возвращает запрос для вставки
// нового параметра-перечисления в базу данных.
func InsertEnumParameter(param *entities.EnumParameter) sq.InsertBuilder {
	data := map[string]any{
		"id":            param.ID,
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
		Where(sq.Eq{"id": id})
}

// SelectParameterIDsByCategory возвращает запрос для получения всех параметров,
// связанных с определенной категорией, включая как числовые, так и перечисляемые параметры.
func SelectParameterIDsByCategory(categoryID uint, parameterType int) sq.SelectBuilder {
	if parameterType == entities.NumericParameterType {
		return psql.
			Select("numeric_parameter_ids").
			From(tableCategories).
			Where(sq.Eq{"id": categoryID})
	}

	if parameterType == entities.EnumParameterType {
		return psql.
			Select("enum_parameter_ids").
			From(tableCategories).
			Where(sq.Eq{"id": categoryID})
	}

	return psql.
		Select("numeric_parameter_ids, enum_parameter_ids").
		From(tableCategories).
		Where(sq.Eq{"id": categoryID})
}

// SelectNumericParameters возвращает запрос для получения всех параметров, без фильтрации по категории,
// включая как числовые параметры.
func SelectNumericParameters(ids []uint) sq.SelectBuilder {
	if len(ids) == 0 {
		return psql.Select(NumericReturningFields).
			From(NumericParametersTable)
	}

	return psql.Select(NumericReturningFields).
		From(NumericParametersTable).
		Where(sq.Eq{"id": ids})
}

// SelectEnumParameters возвращает запрос для получения всех параметров, без фильтрации по категории,
// включая как параметры-перечисления.
func SelectEnumParameters(ids []uint) sq.SelectBuilder {
	if len(ids) == 0 {
		return psql.Select(EnumReturningFields).
			From(EnumParametersTable)
	}

	return psql.Select(EnumReturningFields).
		From(EnumParametersTable).
		Where(sq.Eq{"id": ids})
}

// UpdateNumericParameterInheritance обновляет наследование числовых параметров для указанных категорий.
// Если add равно true, то добавляет параметры к существующему списку, иначе удаляет их.
func UpdateNumericParameterInheritance(parameterIDs []int, add bool, categoriesList []int) sq.UpdateBuilder {
	if add {
		sqlExpr := `(SELECT coalesce(array_agg(DISTINCT x ORDER BY x), '{}')
                     FROM unnest(coalesce(numeric_parameter_ids, '{}') || ?::int[]) AS x)`
		return psql.Update(tableCategories).
			Set("numeric_parameter_ids", sq.Expr(sqlExpr, parameterIDs)).
			Where(sq.Eq{"id": categoriesList})
	}

	sqlExpr := `(SELECT coalesce(array_agg(x ORDER BY x), '{}')
                 FROM unnest(coalesce(numeric_parameter_ids, '{}')) AS x
                 WHERE NOT (x = ANY(?::int[])))`
	return psql.Update(tableCategories).
		Set("numeric_parameter_ids", sq.Expr(sqlExpr, parameterIDs)).
		Where(sq.Eq{"id": categoriesList})
}

// UpdateEnumParameterInheritance обновляет наследование параметров-перечислений для указанных категорий.
// Если add равно true, то добавляет параметры к существующему списку, иначе удаляет их.
func UpdateEnumParameterInheritance(parameterIDs []int, add bool, categoriesList []int) sq.UpdateBuilder {
	if len(parameterIDs) == 0 || len(categoriesList) == 0 {
		return psql.Update(tableCategories).Where(sq.Expr("false"))
	}

	if add {
		sqlExpr := `(SELECT coalesce(array_agg(DISTINCT x ORDER BY x), '{}')
                     FROM unnest(coalesce(enum_parameter_ids, '{}') || ?::int[]) AS x)`
		return psql.Update(tableCategories).
			Set("enum_parameter_ids", sq.Expr(sqlExpr, parameterIDs)).
			Where(sq.Eq{"id": categoriesList})
	}

	sqlExpr := `(SELECT coalesce(array_agg(x ORDER BY x), '{}')
                 FROM unnest(coalesce(enum_parameter_ids, '{}')) AS x
                 WHERE NOT (x = ANY(?::int[])))`
	return psql.Update(tableCategories).
		Set("enum_parameter_ids", sq.Expr(sqlExpr, parameterIDs)).
		Where(sq.Eq{"id": categoriesList})
}
