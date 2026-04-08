package queries

import (
	"Brewery/internal/entities"

	sq "github.com/Masterminds/squirrel"
)

const (
	beersTable     = "beers"
	reviewsTable   = "reviews"
	citiesTable    = "cities"
	countriesTable = "countries"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

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
		"t.name AS type",
		"COALESCE(array_agg(f.name) FILTER (WHERE f.name IS NOT NULL), '{}') AS feats",
	).From(beersTable+" b").
		Join("cities ct ON ct.id = b.city_id").
		Join("countries cntr ON cntr.id = ct.country_id").
		Join("product_categories pc ON pc.id = b.category_id").
		Join("types t ON t.id = b.type_id").
		LeftJoin("beer_features bf ON b.id = bf.beer_id").
		LeftJoin("features f ON f.id = bf.feature_id").
		GroupBy(
			"b.id",
			"ct.name",
			"cntr.name",
			"pc.name",
			"t.name",
		)
}

func SelectBeerByID(id uint) sq.SelectBuilder {
	return FullBeerSelect().Where(sq.Eq{"b.id": id})
}

func SelectBeerByCategoryID(categoryID uint) sq.SelectBuilder {
	return FullBeerSelect().
		Where(sq.Eq{"category_id": categoryID}).
		OrderBy("id DESC")
}

func InsertReview(review entities.Review) sq.InsertBuilder {
	data := map[string]any{
		"body":    review.Body,
		"rating":  review.Rating,
		"beer_id": review.BeerID,
	}

	return psql.Insert("reviews").
		SetMap(data).
		Suffix("RETURNING id")
}

func DeleteBeer(id uint) sq.DeleteBuilder {
	return psql.Delete(beersTable).
		Where(sq.Eq{"id": id})
}

func UpdateBeer(id uint, updates map[string]any) sq.UpdateBuilder {
	return psql.Update(beersTable).
		SetMap(updates).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id")
}
