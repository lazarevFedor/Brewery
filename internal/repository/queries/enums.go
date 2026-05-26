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
// Если entity или field пустые — фильтр по ним не применяется (возвращаются все записи).
func SelectEnumClasses(entity, field string) sq.SelectBuilder {
	q := psql.Select(
		"id",
		"enum_type",
		"entity_name",
		"field_name",
		"unit",
		"is_active",
	).From(enumClassesTable)
	if entity != "" {
		q = q.Where(sq.Eq{"entity_name": entity})
	}
	if field != "" {
		q = q.Where(sq.Eq{"field_name": field})
	}
	return q
}

// SelectEnumClassByID возвращает запрос на получения классов перечислений по имени таблицы и поля.
func SelectEnumClassByID(id uint) sq.SelectBuilder {
	return psql.Select(
		"id",
		"enum_type",
		"entity_name",
		"field_name",
		"unit",
		"is_active",
	).From(enumClassesTable).
		Where(sq.Eq{"id": id})
}

// InsertEnumValue возвращает запрос на вставку значения перечисления.
func InsertEnumValue(enumValue entities.EnumValueRow) sq.InsertBuilder {
	data := map[string]any{
		"enum_class_id": enumValue.EnumClassID,
		"value_raw":     enumValue.ValueRaw,
		"value_type":    enumValue.ValueType,
		"position":      enumValue.Position,
	}

	return psql.
		Insert(enumValuesTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// SelectEnumValues возвращает запрос на получение значений перечисления по имени таблицы, поля и типа значения.
func SelectEnumValues(entity, field string, valueType entities.EnumType) sq.SelectBuilder {
	return psql.
		Select(
			"val.id",
			"val.enum_class_id",
			"val.value_raw",
			"val.value_type",
			"val.position",
		).
		From(enumValuesTable + " val").
		Join("enum_classes cls ON val.enum_class_id = cls.id").
		Where(sq.Eq{"cls.entity_name": entity, "cls.field_name": field, "cls.enum_type": valueType}).
		OrderBy("val.position ASC")
}

// SelectEnumValuesByClassID возвращает запрос на получение значений перечисления по ID класса перечисления.
func SelectEnumValuesByClassID(classID uint) sq.SelectBuilder {
	return psql.
		Select(
			"val.id",
			"val.enum_class_id",
			"val.value_raw",
			"val.value_type",
			"val.position",
		).
		From(enumValuesTable + " val").
		Join("enum_classes cls ON val.enum_class_id = cls.id").
		Where(sq.Eq{"val.enum_class_id": classID}).
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
