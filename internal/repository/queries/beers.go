package queries

import (
	sq "github.com/Masterminds/squirrel"
)

const (
	tableBeers = "beers"
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
		"ct.name AS city",
		"cntr.name AS country",
		"pc.name AS category",
		"t.name AS type",
		"COALESCE(array_agg(f.name) FILTER (WHERE f.name IS NOT NULL), '{}') AS feats",
	).From(tableBeers+" b").
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

func SelectBeerByID(id int) sq.SelectBuilder {
	return FullBeerSelect().Where(sq.Eq{"b.id": id})
}
