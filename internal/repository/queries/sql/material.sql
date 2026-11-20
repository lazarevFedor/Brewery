-- name: InsertMaterial :one
INSERT INTO materials (
    material_id,
    product_id,
    material_amount,
    base
) VALUES (
    sqlc.arg(material_id),
    sqlc.arg(product_id),
    sqlc.arg(material_amount),
    sqlc.arg(base)
) RETURNING
    material_id,
    product_id,
    material_amount,
    base;

-- name: UpdateMaterial :one
UPDATE materials
SET material_amount = sqlc.arg(material_amount),
    base = sqlc.arg(base)
WHERE product_id = sqlc.arg(product_id)
  AND material_id = sqlc.arg(material_id)
RETURNING
    material_id,
    product_id,
    material_amount,
    base;


-- name: DeleteMaterial :execrows
DELETE FROM materials
WHERE material_id = $1
AND product_id = $2;

-- name: GetMaterialsByProductID :many
SELECT 
    m.material_id AS id,
    m.material_amount,
    m.base,
    p.name,
    COALESCE(p.price, 0)::double precision AS price,
    p.is_purchased,
    p.is_final_product,
    p.units,
    p.class_id,
    p.base_id
FROM materials m
JOIN products p ON p.id = m.material_id
WHERE m.product_id = $1;

-- name: GetProductsByMaterial :many
SELECT 
    m.product_id AS id,
    m.material_amount,
    m.base,
    p.name,
    COALESCE(p.price, 0)::double precision AS price,
    p.is_purchased,
    p.is_final_product,
    p.units,
    p.class_id,
    p.base_id,
    c.name AS class_name
FROM materials m
JOIN products p ON p.id = m.product_id
JOIN classes c ON p.class_id = c.id
WHERE m.material_id = $1;


-- name: CopyBaseMaterials :exec
INSERT INTO materials (
    product_id,
    material_id,
    material_amount,
    base
) SELECT 
    sqlc.arg(product_id),
    material_id,
    material_amount,
    base
FROM materials
WHERE materials.product_id = sqlc.arg(base_id)
ON CONFLICT (material_id, product_id) DO NOTHING;

-- name: SyncBaseMaterials :exec
WITH source AS (
    SELECT m.material_id, m.material_amount, m.base
    FROM materials AS m
    WHERE m.product_id = sqlc.arg(base_id)
), removed AS (
    DELETE FROM materials AS current
    WHERE current.product_id = sqlc.arg(product_id)
      AND NOT EXISTS (SELECT 1 FROM source WHERE source.material_id = current.material_id)
)
INSERT INTO materials (product_id, material_id, material_amount, base)
SELECT sqlc.arg(product_id), material_id, material_amount, base FROM source
ON CONFLICT (material_id, product_id) DO UPDATE
SET material_amount = EXCLUDED.material_amount, base = EXCLUDED.base;
