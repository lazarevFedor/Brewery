INSERT INTO cities (name, country_id)
VALUES ($1, $2)
ON CONFLICT (name, country_id) DO UPDATE
SET name = EXCLUDED.name
RETURNING id;