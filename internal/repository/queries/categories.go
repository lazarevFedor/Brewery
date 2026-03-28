package queries

import (
	"Brewery/internal/entities"
	sq "github.com/Masterminds/squirrel"
)

const (
	tableCategories     = "product_categories"
	categoryIDCol       = "id"
	categoryNameCol     = "name"
	categoryParentIDCol = "parent_id"
)

func FullCategorySelect() sq.SelectBuilder {
	return psql.Select(
		categoryIDCol,
		categoryNameCol,
		categoryParentIDCol,
	).From(tableCategories)
}

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

func SelectCategoryByID(id uint) sq.SelectBuilder {
	return FullBeerSelect().Where(sq.Eq{categoryIDCol: id})
}

func UpdateCategory(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.Update(tableCategories).
		SetMap(updates).
		Where(sq.Eq{categoryIDCol: id})
}

func DeleteCategory(id uint) sq.DeleteBuilder {
	return psql.Delete(tableCategories).Where(sq.Eq{categoryIDCol: id})
}
