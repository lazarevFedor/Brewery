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


-- name: DeleteMaterial :exec
DELETE FROM materials
WHERE material_id = $1
AND product_id = $2;

-- name: GetMaterialsByProductID :many
SELECT 
    m.material_id AS id,
    m.material_amount,
    m.base,
    p.name,
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
    p.units,
    p.class_id,
    p.base_id
FROM materials m
JOIN products p ON p.id = m.product_id
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

-- name: DeleteProductsMaterials :exec
DELETE FROM materials
WHERE product_id = $1;
