// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	// Константы с именами таблиц, используемых в запросах
	beersTable        = "beers"
	citiesTable       = "cities"
	countriesTable    = "countries"
	featuresTable     = "features"
	beerFeaturesTable = "beer_features"
)

// Exists возвращает pапрос на проверку наличия сущности пива по id
func Exists() sq.SelectBuilder {
	return psql.Select("EXISTS(SELECT 1 FROM beers WHERE id = $1)")
}

// FullBeerSelect возвращает базовый запрос для получения полной информации о пиве, включая его характеристики, город и страну производства, категорию и особенности.
func FullBeerSelect() sq.SelectBuilder {
	return psql.Select(
		"b.id",
		"b.name",
		"b.description",
		"b.abv",
		"b.ibu",
		"b.amount",
		"b.units",
		"ct.name AS city",
		"cntr.name AS country",
		"pc.name AS category",
		"COALESCE(array_agg(f.name) FILTER (WHERE f.name IS NOT NULL), '{}') AS feats",
	).From(beersTable+" b").
		Join("cities ct ON ct.id = b.city_id").
		Join("countries cntr ON cntr.id = ct.country_id").
		Join("product_categories pc ON pc.id = b.category_id").
		LeftJoin("beer_features bf ON b.id = bf.beer_id").
		LeftJoin("features f ON f.id = bf.feature_id").
		GroupBy(
			"b.id",
			"ct.name",
			"cntr.name",
			"pc.name",
		)
}

// SelectBeerByID возвращает запрос для получения информации о пиве по его ID.
func SelectBeerByID(id uint) sq.SelectBuilder {
	return FullBeerSelect().Where(sq.Eq{"b.id": id})
}

// SelectBeerByCategoryID возвращает запрос для получения списка пива, принадлежащего к определенной категории, с сортировкой по убыванию ID.
func SelectBeerByCategoryID(categoryID uint) sq.SelectBuilder {
	return FullBeerSelect().
		Where(sq.Eq{"category_id": categoryID}).
		OrderBy("id DESC")
}


func FilterBeers(filters []*entities.FilterParameter, categoryID uint) sq.SelectBuilder {
	query := FullBeerSelect()
	if len(filters) != 0 {
		for _, filter := range filters {
			field := filter.FieldName
			val := filter.Value
			var pred any
			
			oper := filter.Operation
			switch oper {
			case entities.OpEqual:
				pred = sq.Eq{field: val}
			case entities.OpGreater:
				pred = sq.Gt{field: val}
			case entities.OpGreaterEqual:
				pred = sq.GtOrEq{field: val}
			case entities.OpLess:
				pred = sq.Lt{field: val}
			case entities.OpLessEqual:
				pred = sq.LtOrEq{field: val}
			case entities.OpNotEqual:
				pred = sq.NotEq{field: val}
			}
			
			query = query.Where(pred)
		}
	}

	if categoryID != 0 {
		query = query.Where(sq.Eq{"category_id": categoryID})
	}
	return query
}

// DeleteBeer возвращает запрос для удаления пива из таблицы beers по его ID.
func DeleteBeer(id uint) sq.DeleteBuilder {
	return psql.
		Delete(beersTable).
		Where(sq.Eq{"id": id})
}

// UpdateBeer возвращает запрос для обновления информации о пиве в таблице beers по его ID с использованием данных из переданной карты обновлений.
func UpdateBeer(id uint, updates map[string]any) sq.UpdateBuilder {
	returnColumns := "id, name, description, abv, ibu, amount, units, city_id, category_id"
	return psql.
		Update(beersTable).
		SetMap(updates).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + returnColumns)
}

// SelectOrInsertCountry возвращает запрос для вставки новой страны в таблицу countries, если страна с таким именем еще не существует, или возвращает ID существующей страны, если она уже есть. Запрос использует конструкцию ON CONFLICT для обработки конфликтов по имени страны.
func SelectOrInsertCountry(name string) sq.InsertBuilder {
	data := map[string]any{
		"name": name,
	}

	return psql.
		Insert(countriesTable).
		SetMap(data).
		Suffix("ON CONFLICT (name) DO UPDATE " +
			"SET name = EXCLUDED.name " +
			"RETURNING id")
}

// SelectOrInsertCity возвращает запрос для вставки нового города в таблицу cities, если город с таким именем и идентификатором страны еще не существует, или возвращает ID существующего города, если он уже есть. Запрос использует конструкцию ON CONFLICT для обработки конфликтов по имени города и идентификатору страны.
func SelectOrInsertCity(cityName string, countryID uint) sq.InsertBuilder {
	data := map[string]any{
		"name":       cityName,
		"country_id": countryID,
	}

	return psql.
		Insert(citiesTable).
		SetMap(data).
		Suffix("ON CONFLICT (name, country_id) DO UPDATE " +
			"SET name = EXCLUDED.name " +
			"RETURNING id")
}

// SelectOrInsertFeature возвращает запрос для вставки новой особенности в таблицу features, если особенность с таким именем еще не существует, или возвращает ID существующей особенности, если она уже есть. Запрос использует конструкцию ON CONFLICT для обработки конфликтов по имени особенности.
func SelectOrInsertFeature(featName string) sq.InsertBuilder {
	data := map[string]any{
		"name": featName,
	}

	return psql.
		Insert(featuresTable).
		SetMap(data).
		Suffix("ON CONFLICT (name) DO UPDATE " +
			"SET name = EXCLUDED.name " +
			"RETURNING id")
}

// ConnectBeerAndFeature возвращает запрос для вставки новой связи между пивом и особенностью в таблицу beer_features, если такая связь еще не существует, или ничего не делает, если связь уже есть. Запрос использует конструкцию ON CONFLICT для обработки конфликтов по идентификаторам пива и особенности.
func ConnectBeerAndFeature(featID, beerID uint) sq.InsertBuilder {
	data := map[string]any{
		"beer_id":    beerID,
		"feature_id": featID,
	}

	return psql.
		Insert(beerFeaturesTable).
		SetMap(data).
		Suffix("ON CONFLICT DO NOTHING")
}

func DisconnectBeerAndFeature(beerID uint) sq.DeleteBuilder {
	return psql.
		Delete(beerFeaturesTable).
		Where(sq.Eq{"beer_id": beerID})
}

// SelectBeersFeature возвращает запрос для получения списка особенностей пива
func SelectBeersFeature(beerID uint) sq.SelectBuilder {
	return psql.
		Select("name").
		From(beerFeaturesTable + " bf").
		Join(featuresTable + " f ON f.id = bf.feature_id").
		Where(sq.Eq{"beer_id": beerID})
}

// InsertBeer возвращает запрос для вставки нового пива в таблицу beers с использованием данных из переданной структуры beer и идентификаторов города и категории. Запрос возвращает ID вставленного пива.
func InsertBeer(beer entities.Beer, cityID, categoryID uint) sq.InsertBuilder {
	data := map[string]any{
		"name":        beer.Name,
		"description": beer.Description,
		"abv":         beer.ABV,
		"ibu":         beer.IBU,
		"amount":      beer.Amount,
		"units":       beer.Unit,
		"city_id":     cityID,
		"category_id": categoryID,
	}

	return psql.
		Insert(beersTable).
		SetMap(data).
		Suffix("RETURNING *")
}

// SelectCityNameByID возвращает запрос для получения названия города по его ID из таблицы cities.
func SelectCityNameByID(cityID uint) sq.SelectBuilder {
	return psql.
		Select("name").
		From(citiesTable).
		Where(sq.Eq{"id": cityID})
}
