package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	enumParametersTable = "enum_parameters"
	enumParameterFields = "id, enum_type, field_name, unit, is_active, inheritable"
)

// InsertEnumParameter возвращает запрос на вставку перечисляемого параметра.
func InsertEnumParameter(parameter entities.EnumClass) sq.InsertBuilder {
	data := map[string]any{
		"enum_type":   parameter.Type,
		"field_name":  parameter.FieldName,
		"is_active":   parameter.IsActive,
		"inheritable": parameter.Inheritable,
	}

	if parameter.Unit != "" {
		data["unit"] = parameter.Unit
	}

	return psql.
		Insert(enumParametersTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// UpdateEnumParameter возвращает запрос на обновление перечисляемого параметра.
func UpdateEnumParameter(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(enumParametersTable).
		SetMap(updates).
		Where(sq.Eq{"id": id})
}

// DeleteEnumParameter возвращает запрос на удаление перечисляемого параметра.
func DeleteEnumParameter(id uint) sq.DeleteBuilder {
	return psql.
		Delete(enumParametersTable).
		Where(sq.Eq{"id": id})
}

// SelectEnumParameters возвращает запрос на получение перечисляемых параметров по имени поля.
func SelectEnumParameters(field string) sq.SelectBuilder {
	return psql.Select(enumParameterFields).From(enumParametersTable).
		Where(sq.Eq{"field_name": field})
}

// SelectEnumParametersByIDs возвращает перечисляемые параметры; пустой список означает отсутствие фильтра.
func SelectEnumParametersByIDs(ids []int64) sq.SelectBuilder {
	query := psql.Select(enumParameterFields).From(enumParametersTable)
	if len(ids) > 0 {
		query = query.Where(sq.Eq{"id": ids})
	}
	return query
}

// SelectEnumParameterByID возвращает перечисляемый параметр по ID.
func SelectEnumParameterByID(id uint) sq.SelectBuilder {
	return psql.Select(enumParameterFields).From(enumParametersTable).
		Where(sq.Eq{"id": id})
}
