// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

const (
	aggregatesTable = "aggregates"

	aggregateReturningFields = "id, name, description, numeric_parameter_ids, enum_parameter_ids"
)

// InsertAggregate возвращает запрос для вставки нового агрегата в таблицу aggregates
// с использованием данных из предоставленной структуры Aggregate
// и возвращает ID вставленного агрегата.
func InsertAggregate(aggregate *entities.Aggregate) sq.InsertBuilder {
	data := map[string]any{
		"name":                  aggregate.Name,
		"description":           aggregate.Description,
		"numeric_parameter_ids": pq.Array(aggregate.NumericParameters),
		"enum_parameter_ids":    pq.Array(aggregate.EnumParameters),
	}
	return psql.Insert(aggregatesTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// UpdateAggregate возвращает запрос для обновления агрегата в таблице aggregates по его ID
// с использованием предоставленных обновлений и возвращает обновленный агрегат.
func UpdateAggregate(id int64, updates entities.AggregateUpdate) sq.UpdateBuilder {
	query := psql.Update(aggregatesTable)
	if updates.Name != nil {
		query = query.Set("name", *updates.Name)
	}
	if updates.Description != nil {
		query = query.Set("description", *updates.Description)
	}
	if updates.RangeParamIDs != nil {
		query = query.Set("numeric_parameter_ids", updates.RangeParamIDs)
	}
	if updates.EnumParamIDs != nil {
		query = query.Set("enum_parameter_ids", updates.EnumParamIDs)
	}
	return query.
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + aggregateReturningFields)
}

// DeleteAggregate возвращает запрос для удаления агрегата
// из таблицы aggregates по его ID и возвращает удаленный агрегат.
func DeleteAggregate(id int64) sq.DeleteBuilder {
	return psql.
		Delete(aggregatesTable).
		Where(sq.Eq{"id": id})
}

// GetAggregates возвращает запрос для получения всех агрегатов,
// соответствующих заданному имени, из таблицы aggregates.
func GetAggregates(name string) sq.SelectBuilder {
	query := psql.
		Select(aggregateReturningFields).
		From(aggregatesTable)

	if name != "" {
		query = query.Where(sq.Eq{"name": name}).OrderBy("name")
	}

	return query
}

// ApplyAggregate дополняет параметры класса параметрами агрегата.
func ApplyAggregate(classID int64, id int64) sq.UpdateBuilder {
	return psql.
		Update(tableClasses).
		Set("numeric_parameter_ids", sq.Expr(`CASE WHEN cardinality(a.numeric_parameter_ids) > 0 THEN ARRAY(
			SELECT unnest(classes.numeric_parameter_ids)
			UNION SELECT unnest(a.numeric_parameter_ids)
			ORDER BY 1
		) ELSE classes.numeric_parameter_ids END`)).
		Set("enum_parameter_ids", sq.Expr(`CASE WHEN cardinality(a.enum_parameter_ids) > 0 THEN ARRAY(
			SELECT unnest(classes.enum_parameter_ids)
			UNION SELECT unnest(a.enum_parameter_ids)
			ORDER BY 1
		) ELSE classes.enum_parameter_ids END`)).
		From(aggregatesTable + " a").
		Where(sq.Eq{"classes.id": classID, "a.id": id})
}
