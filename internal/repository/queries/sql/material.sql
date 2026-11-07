-- name: InsertMaterial :one
INSERT INTO materials (
    material_id,
    product_id,
    material_amount
) VALUES (
    sqlc.arg(material_id),
    sqlc.arg(product_id),
    sqlc.arg(material_amount)
) RETURNING
    material_id,
    product_id,
    material_amount;

-- name: UpdateMaterial :one
UPDATE materials
SET 
    material_id = COALESCE(sqlc.narg(material_id), material_id),
    product_id = COALESCE(sqlc.narg(product_id), product_id),
    material_amount = COALESCE(sqlc.narg(material_amount), material_amount)
RETURNING
    material_id,
    product_id,
    material_amount;


-- name: DeleteMaterial :exec
DELETE FROM materials
WHERE material_id = $1
AND product_id = $2;

-- name: GetMaterialsByProductID :many
SELECT 
    m.material_id AS id,
    m.material_amount,
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
    p.name,
    p.units,
    p.class_id,
    p.base_id
FROM materials m
JOIN products p ON p.id = m.product_id
WHERE m.material_id = $1;