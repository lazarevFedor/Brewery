// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	rangeParametersTable = "range_parameters"

	rangeParameterFields = "id, min_val, max_val, field_name, inheritable, unit"
)

// InsertRangeParameter возвращает запрос для вставки
// нового числового параметра в базу данных.
func InsertRangeParameter(param *entities.RangeParameter) sq.InsertBuilder {
	data := map[string]any{
		"min_val":     param.MinValue,
		"max_val":     param.MaxValue,
		"field_name":  param.FieldName,
		"inheritable": param.Inheritable,
		"unit":        param.Unit,
	}

	return psql.
		Insert(rangeParametersTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// UpdateRangeParameter возвращает запрос для обновления
// существующего числового параметра в базе данных.
func UpdateRangeParameter(id int64, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(rangeParametersTable).
		SetMap(updates).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + rangeParameterFields)
}

// DeleteRangeParameter возвращает запрос для
// удаления числового параметра из базы данных.
func DeleteRangeParameter(id int64) sq.DeleteBuilder {
	return psql.
		Delete(rangeParametersTable).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + rangeParameterFields)
}

// SelectRangeParameters выбирает числовые параметры; пустой список ID отключает фильтрацию.
func SelectRangeParameters(ids []int64) sq.SelectBuilder {
	query := psql.
		Select(rangeParameterFields).
		From(rangeParametersTable)

	if len(ids) > 0 {
		query = query.Where(sq.Eq{"id": ids})
	}
	return query
}

// SelectRangeParametersByField возвращает параметры по имени поля; пустое имя отключает фильтрацию.
func SelectRangeParametersByField(fieldName string) sq.SelectBuilder {
	query := SelectRangeParameters(nil).OrderBy("id ASC")
	if fieldName != "" {
		query = query.Where(sq.Eq{"field_name": fieldName})
	}
	return query
}
