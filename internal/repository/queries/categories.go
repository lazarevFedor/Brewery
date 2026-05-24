// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	// Константы с именами таблиц, используемых в запросах
	tableCategories     = "product_categories"
	categoryIDCol       = "id"
	categoryNameCol     = "name"
	categoryParentIDCol = "parent_id"
)

// FullCategorySelect возвращает базовый запрос для получения всех полей категории, включая обработку NULL для parent_id
func FullCategorySelect() sq.SelectBuilder {
	return psql.Select(
		categoryIDCol,
		categoryNameCol,
		"COALESCE(parent_id, 0)",
	).From(tableCategories)
}

// CategoryInsert возвращает запрос для вставки новой категории, учитывая необязательное поле parent_id
func CategoryInsert(category entities.ProductCategory) sq.InsertBuilder {
	data := map[string]any{
		categoryNameCol: category.Name,
	}
	if category.ParentID != 0 {
		data[categoryParentIDCol] = category.ParentID
	}

	return psql.Insert(tableCategories).
		SetMap(data).
		Suffix("RETURNING id")
}

// SelectCategoryByID возвращает запрос для получения категории по её ID, включая обработку NULL для parent_id
func SelectCategoryByID(id uint) sq.SelectBuilder {
	return FullCategorySelect().Where(sq.Eq{categoryIDCol: id})
}

// SelectCategoryByName возвращает запрос для получения категории по её имени, включая обработку NULL для parent_id
func SelectCategoryByName(name string) sq.SelectBuilder {
	return psql.Select(categoryIDCol).
		From(tableCategories).
		Where(sq.Eq{categoryNameCol: name})
}

// UpdateCategory возвращает запрос для обновления категории по её ID, учитывая возможность обновления parent_id
func UpdateCategory(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.Update(tableCategories).
		SetMap(updates).
		Where(sq.Eq{categoryIDCol: id})
}

// DeleteCategory возвращает запрос для удаления категории по её ID
func DeleteCategory(id uint) sq.DeleteBuilder {
	return psql.Delete(tableCategories).Where(sq.Eq{categoryIDCol: id})
}

// SelectChildrenCategories возвращает запрос для получения всех дочерних категорий по ID родительской категории
func SelectChildrenCategories(id uint) sq.SelectBuilder {
	return psql.Select(categoryIDCol).
		From(tableCategories).
		Where(sq.Eq{categoryParentIDCol: id})
}
