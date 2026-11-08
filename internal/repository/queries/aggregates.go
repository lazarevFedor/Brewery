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
func UpdateAggregate(id int64, updates map[string]any) sq.UpdateBuilder {
	if updates == nil {
		return psql.Update(aggregatesTable).
			SetMap(updates).
			Where(sq.Eq{"id": id}).
			Suffix("RETURNING " + aggregateReturningFields)
	}

	return psql.Update(aggregatesTable).
		SetMap(updates).
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

// ApplyAggregate возвращает запрос для применения агрегата к классу,
// обновляя поля numeric_parameter_ids и enum_parameter_ids в таблице product_categories
// на значения из агрегата, соответствующего заданному ID.
func ApplyAggregate(classID int64, id int64) sq.UpdateBuilder {
	return psql.
		Update(tableClasses).
		Set("numeric_parameter_ids", sq.Expr("a.numeric_parameter_ids")).
		Set("enum_parameter_ids", sq.Expr("a.enum_parameter_ids")).
		From(aggregatesTable + " a").
		Where(sq.Eq{"classes.id": classID, "a.id": id})
}
