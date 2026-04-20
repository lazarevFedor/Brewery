package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	enumClassesTable = "enum_classes"
	enumValuesTable  = "enum_values"
)

func InsertEnumClass(enumClass entities.EnumClass) sq.InsertBuilder {
	data := map[string]any{
		"enum_type":   enumClass.Type,
		"entity_name": enumClass.EntityName,
		"field_name":  enumClass.FieldName,
	}

	return psql.
		Insert(enumClassesTable).
		SetMap(data).
		Suffix("RETURNING id")
}

func UpdateEnumClass(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(enumClassesTable).
		SetMap(updates).
		Where(sq.Eq{"id": id})
}

func DeleteEnumClass(id uint) sq.DeleteBuilder {
	return psql.
		Delete(enumClassesTable).
		Where(sq.Eq{"id": id})
}

func SelectEnumClasses(entity, field string) sq.SelectBuilder {
	return psql.Select(
		"id",
		"enum_type",
		"entity_name",
		"field_name",
	).From(enumClassesTable).
		Where(sq.Eq{"entity_name": entity, "field_name": field})
}
