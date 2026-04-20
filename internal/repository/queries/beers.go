// Package queries содержит функции для сборки запросов к базе данных
package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	beersTable        = "beers"
	reviewsTable      = "reviews"
	citiesTable       = "cities"
	countriesTable    = "countries"
	featuresTable     = "features"
	beerFeaturesTable = "beer_features"
)

// FullBeerSelect возвращает базовый запрос для получения полной информации о пиве, включая его характеристики, город и страну производства, категорию и особенности.
func FullBeerSelect() sq.SelectBuilder {
	return psql.Select(
		"b.id",
		"b.name",
		"b.rating",
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

// InsertReview возвращает запрос для вставки нового отзыва в таблицу reviews и возвращает ID вставленного отзыва.
func InsertReview(review entities.Review) sq.InsertBuilder {
	data := map[string]any{
		"body":    review.Body,
		"rating":  review.Rating,
		"beer_id": review.BeerID,
	}

	return psql.
		Insert(reviewsTable).
		SetMap(data).
		Suffix("RETURNING id")
}

// DeleteBeer возвращает запрос для удаления пива из таблицы beers по его ID.
func DeleteBeer(id uint) sq.DeleteBuilder {
	return psql.
		Delete(beersTable).
		Where(sq.Eq{"id": id})
}

// UpdateBeer возвращает запрос для обновления информации о пиве в таблице beers по его ID с использованием данных из переданной карты обновлений.
func UpdateBeer(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.
		Update(beersTable).
		SetMap(updates).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id")
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

// SelectOrInsertBeerFeature возвращает запрос для вставки новой связи между пивом и особенностью в таблицу beer_features, если такая связь еще не существует, или ничего не делает, если связь уже есть. Запрос использует конструкцию ON CONFLICT для обработки конфликтов по идентификаторам пива и особенности.
func SelectOrInsertBeerFeature(featID, beerID uint) sq.InsertBuilder {
	data := map[string]any{
		"beer_id":    beerID,
		"feature_id": featID,
	}

	return psql.
		Insert(beerFeaturesTable).
		SetMap(data).
		Suffix("ON CONFLICT DO NOTHING")
}

// InsertBeer возвращает запрос для вставки нового пива в таблицу beers с использованием данных из переданной структуры beer и идентификаторов города и категории. Запрос возвращает ID вставленного пива.
func InsertBeer(beer entities.Beer, cityID, categoryID uint) sq.InsertBuilder {
	data := map[string]any{
		"name":        beer.Name,
		"rating":      beer.Rating,
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
		Suffix("RETURNING id")
}
