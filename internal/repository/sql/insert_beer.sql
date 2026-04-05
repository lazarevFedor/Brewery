INSERT INTO beers (
    name,
    rating,
    description,
    abv,
    ibu,
    amount,
    units,
    type_id,
    city_id,
    category_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;