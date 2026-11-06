// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	// Константы с именами таблиц, используемых в запросах
	tableClasses     = "classes"
	classIDCol       = "id"
	classNameCol     = "name"
	classParentIDCol = "parent_id"
)

// FullClassSelect возвращает базовый запрос для получения всех полей категории, включая обработку NULL для parent_id
func FullClassSelect() sq.SelectBuilder {
	return psql.Select(
		classIDCol,
		classNameCol,
		"COALESCE(parent_id, 0)",
	).From(tableClasses)
}

// ClassInsert возвращает запрос для вставки новой категории, учитывая необязательное поле parent_id
func ClassInsert(class entities.Class) sq.InsertBuilder {
	data := map[string]any{
		classNameCol: class.Name,
	}
	if class.ParentID != 0 {
		data[classParentIDCol] = class.ParentID
	}

	return psql.Insert(tableCategories).
		SetMap(data).
		Suffix("RETURNING id")
}

// SelectClassByID возвращает запрос для получения категории по её ID, включая обработку NULL для parent_id
func SelectClassByID(id uint) sq.SelectBuilder {
	return FullClassSelect().Where(sq.Eq{classIDCol: id})
}

// SelectClassByName возвращает запрос для получения категории по её имени, включая обработку NULL для parent_id
func SelectClassByName(name string) sq.SelectBuilder {
	return psql.Select(classIDCol).
		From(tableCategories).
		Where(sq.Eq{classNameCol: name})
}

// UpdateClass возвращает запрос для обновления категории по её ID, учитывая возможность обновления parent_id
func UpdateClass(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.Update(tableCategories).
		SetMap(updates).
		Where(sq.Eq{classIDCol: id})
}

// DeleteClass возвращает запрос для удаления категории по её ID
func DeleteClass(id uint) sq.DeleteBuilder {
	return psql.Delete(tableCategories).Where(sq.Eq{classIDCol: id})
}

// SelectChildrenCategories возвращает запрос для получения всех дочерних категорий по ID родительской категории
func SelectChildrenClasses(id uint) sq.SelectBuilder {
	return psql.Select(classIDCol).
		From(tableCategories).
		Where(sq.Eq{classParentIDCol: id})
}
