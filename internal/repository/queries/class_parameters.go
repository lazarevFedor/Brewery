package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

// SelectParameterIDsByClass возвращает запрос для получения всех параметров,
// связанных с указанным классом, включая как числовые, так и перечисляемые параметры.
func SelectParameterIDsByClass(classID int64, parameterType int) sq.SelectBuilder {
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
	if classID != 0 {
		query = query.Where(sq.Eq{"id": classID})
	}

	return query
}

// AddRangeParametersToClasses возвращает запрос для добавления числовых параметров в классы
func AddRangeParametersToClasses(parameterIDs []int64, classIDs []int64) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"add_numeric_parameters_to_categories(?::bigint[], ?::bigint[])",
				pq.Array(parameterIDs),
				pq.Array(classIDs),
			),
		)
}

// AddEnumParametersToClasses возвращает запрос для добавления параметров-перечислений в классы
func AddEnumParametersToClasses(parameterIDs []int64, classIDs []int64) sq.SelectBuilder {
	return psql.
		Select().
		Column(
			sq.Expr(
				"add_enum_parameters_to_categories(?::bigint[], ?::bigint[])",
				pq.Array(parameterIDs),
				pq.Array(classIDs),
			),
		)
}

// InheritParametersToChildren возвращает запрос на наследование параметров дочерними классами
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

// SelectClassesWithRangeParameter возвращает запрос для получения всех классов, содержащих числовой параметр
func SelectClassesWithRangeParameter(parameterID int64) sq.SelectBuilder {
	return psql.
		Select("id").
		From(tableClasses).
		Where(sq.Expr("numeric_parameter_ids @> ARRAY[?::bigint]", parameterID))
}

// SelectClassesWithEnumParameter возвращает запрос для получения всех классов, содержащих параметр-перечисление
func SelectClassesWithEnumParameter(parameterID int64) sq.SelectBuilder {
	return psql.
		Select("id").
		From(tableClasses).
		Where(sq.Expr("enum_parameter_ids @> ARRAY[?::bigint]", parameterID))
}

// RemoveRangeParametersFromDescendants возвращает запрос для удаления числовых параметров из потомков классов
func RemoveRangeParametersFromDescendants(parameterIDs []int64, parentIDs []int64) sq.SelectBuilder {
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

// RemoveEnumParametersFromDescendants возвращает запрос для удаления параметров-перечислений из потомков классов
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

// RemoveRangeParametersFromClasses убирает связи числовых параметров с указанными классами.
func RemoveRangeParametersFromClasses(parameterIDs, classIDs []int64) sq.SelectBuilder {
	return psql.Select().Column(sq.Expr(
		"remove_numeric_parameters_from_categories(?::bigint[], ?::bigint[])",
		pq.Array(parameterIDs), pq.Array(classIDs),
	))
}

// RemoveEnumParametersFromClasses убирает связи перечисляемых параметров с указанными классами.
func RemoveEnumParametersFromClasses(parameterIDs, classIDs []int64) sq.SelectBuilder {
	return psql.Select().Column(sq.Expr(
		"remove_enum_parameters_from_categories(?::bigint[], ?::bigint[])",
		pq.Array(parameterIDs), pq.Array(classIDs),
	))
}

// SelectInheritableRangeParameterIDs возвращает наследуемые параметры из заданного списка.
func SelectInheritableRangeParameterIDs(ids []int64) sq.SelectBuilder {
	return psql.Select("id").From(rangeParametersTable).Where(sq.Eq{"id": ids, "inheritable": true})
}

// SelectInheritableEnumParameterIDs возвращает наследуемые параметры из заданного списка.
func SelectInheritableEnumParameterIDs(ids []int64) sq.SelectBuilder {
	return psql.Select("id").From(enumParametersTable).Where(sq.Eq{"id": ids, "inheritable": true})
}
