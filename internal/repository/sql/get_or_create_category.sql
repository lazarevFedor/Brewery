INSERT INTO product_categories (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE
    SET name = EXCLUDED.name
RETURNING id;