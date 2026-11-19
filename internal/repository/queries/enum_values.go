package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const enumValuesTable = "parameter_values"

// InsertEnumValue возвращает запрос на вставку значения перечисления.
func InsertEnumValue(enumValue entities.ParameterValueRow) sq.InsertBuilder {
	data := map[string]any{
		"enum_parameter_id": enumValue.EnumClassID,
		"value_raw":         enumValue.ValueRaw,
		"position":          enumValue.Position,
	}

	return psql.
		Insert(enumValuesTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// SelectEnumValues возвращает запрос на получение значений перечисления по имени таблицы, поля и типа значения.
func SelectEnumValues(field string, valueType entities.EnumType) sq.SelectBuilder {
	return psql.
		Select(
			"val.id",
			"val.enum_parameter_id",
			"val.value_raw",
			"param.enum_type AS value_type",
			"val.position",
		).
		From(enumValuesTable + " val").
		Join(enumParametersTable + " param ON val.enum_parameter_id = param.id").
		Where(sq.Eq{"param.field_name": field, "param.enum_type": valueType}).
		OrderBy("val.position ASC")
}

// SelectEnumValuesByParameterID возвращает запрос на получение значений перечисления по ID перечисляемого параметра.
func SelectEnumValuesByParameterID(parameterID uint) sq.SelectBuilder {
	return psql.
		Select(
			"val.id",
			"val.enum_parameter_id",
			"val.value_raw",
			"param.enum_type AS value_type",
			"val.position",
		).
		From(enumValuesTable + " val").
		Join(enumParametersTable + " param ON val.enum_parameter_id = param.id").
		Where(sq.Eq{"val.enum_parameter_id": parameterID}).
		OrderBy("val.position ASC")
}

// UpdateEnumValue возвращает запрос на обновление значения перечисления.
func UpdateEnumValue(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(enumValuesTable).
		SetMap(updates).
		Where(sq.Eq{"id": id})
}

// DeleteEnumValue возвращает запрос на удаление значения перечисления.
func DeleteEnumValue(id uint) sq.DeleteBuilder {
	return psql.
		Delete(enumValuesTable).
		Where(sq.Eq{"id": id})
}
