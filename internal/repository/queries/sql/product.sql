-- name: InsertProduct :one
INSERT INTO products(
    name,
    amount,
    units,
    base_id
) VALUES (
    sqlc.arg(name),
    sqlc.arg(amount),
    sqlc.arg(units),
    sqlc.arg(base_id)
) RETURNING 
    id,
    name,
    amount,
    class_id,
    units,
    base_id;


-- name: GetProductByID :one
SELECT 
    p.id, 
    p.name, 
    p.amount,
    p.units,
    p.base_id,
    c.id AS class_id,   
    c.name AS class_name,
    c.parent_id AS class_parent_id,
    c.numeric_parameter_ids AS class_num_params,
    c.enum_parameter_ids AS class_enum_params
FROM products p
JOIN classes c ON p.class_id = c.id
WHERE p.id = $1;
 

-- name: GetProducts :many
SELECT
    id,
    name, 
    amount, 
    units,
    class_id, 
    base_id
FROM products
LIMIT $1
OFFSET $2;
 
-- name: UpdateProduct :one
UPDATE products
SET
    name = COALESCE(sqlc.narg(name), name),
    amount = COALESCE(sqlc.narg(amount), amount),
    units = COALESCE(sqlc.narg(units), units),
    class_id = COALESCE(sqlc.narg(class_id), class_id),
    base_id = COALESCE(sqlc.narg(base_id), base_id)
WHERE id = sqlc.arg(id)
RETURNING 
    id,
    name,
    amount,
    units,
    base_id;
 
-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

-- name: GetProductByClassID :many
SELECT
    id,
    name, 
    amount, 
    units,
    class_id, 
    base_id
FROM products
WHERE class_id = $1;
