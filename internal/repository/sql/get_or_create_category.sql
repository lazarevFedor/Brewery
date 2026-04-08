WITH ins AS (
    INSERT INTO product_categories (name)
    VALUES ($1)
    ON CONFLICT (name) DO NOTHING
    RETURNING id
)
SELECT id FROM ins
UNION ALL
SELECT id FROM product_categories WHERE name = $1
LIMIT 1;
