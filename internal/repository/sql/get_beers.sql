SELECT  b.id,  
        b.name,
        b.rating,
        b.description,
        b.abv,
        b.ibu,
        ct.name AS city,
        cntr.name AS country,
        pc.name AS category,
        t.name AS type,
        COALESCE(array_agg(f.name) FILTER (WHERE f.name IS NOT NULL), '{}') AS feats
FROM beers b
JOIN cities ct ON ct.id = b.city_id
JOIN countries cntr ON cntr.id = ct.country_id
JOIN product_categories pc ON pc.id = b.category_id
JOIN types t ON t.id = b.type_id
LEFT JOIN beer_features bf ON b.id = bf.beer_id
LEFT JOIN features f ON f.id = bf.feature_id
GROUP BY 
    b.id, ct.name, cntr.name, pc.name, t.name;