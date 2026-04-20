package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	enumClassesTable = "enum_classes"
	enumValuesTable  = "enum_values"
)

// InsertEnumClass возвращает запрос на вставку класса перечисления.
func InsertEnumClass(enumClass entities.EnumClass) sq.InsertBuilder {
	data := map[string]any{
		"enum_type":   enumClass.Type,
		"entity_name": enumClass.EntityName,
		"field_name":  enumClass.FieldName,
		"is_active":   enumClass.IsActive,
	}

	if enumClass.Unit != "" {
		data["unit"] = enumClass.Unit
	}

	return psql.
		Insert(enumClassesTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// UpdateEnumClass возращает запрос на обновление класса перечисления.
func UpdateEnumClass(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(enumClassesTable).
		SetMap(updates).
		Where(sq.Eq{"id": id})
}

// DeleteEnumClass возвращает запрос на удаление класса перечисления.
func DeleteEnumClass(id uint) sq.DeleteBuilder {
	return psql.
		Delete(enumClassesTable).
		Where(sq.Eq{"id": id})
}

// SelectEnumClasses возвращает запрос на получения классов перечислений по имени таблицы и поля.
func SelectEnumClasses(entity, field string) sq.SelectBuilder {
	return psql.Select(
		"id",
		"enum_type",
		"entity_name",
		"field_name",
		"unit",
		"is_active",
	).From(enumClassesTable).
		Where(sq.Eq{"entity_name": entity, "field_name": field})
}
