-- name: InsertProduct :one
INSERT INTO products(
    name,
    amount,
    units,
    class_id,
    base_id,
    price,
    is_purchased,
    is_final_product
) VALUES (
    sqlc.arg(name),
    sqlc.arg(amount),
    sqlc.arg(units),
    sqlc.arg(class_id),
    sqlc.arg(base_id),
    sqlc.arg(price),
    sqlc.arg(is_purchased),
    sqlc.arg(is_final_product)
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
    COALESCE(p.price, 0)::double precision AS price,
    p.is_purchased,
    p.is_final_product,
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
    p.id,
    p.name,
    p.amount,
    p.units,
    p.base_id,
    COALESCE(p.price, 0)::double precision AS price,
    p.is_purchased,
    p.is_final_product,
    c.id AS class_id,
    c.name AS class_name,
    c.parent_id AS class_parent_id,
    c.numeric_parameter_ids AS class_num_params,
    c.enum_parameter_ids AS class_enum_params
FROM products p
JOIN classes c ON p.class_id = c.id
ORDER BY p.id
LIMIT $1
OFFSET $2;
 
-- name: UpdateProduct :one
UPDATE products
SET
    name = COALESCE(sqlc.narg(name), name),
    amount = COALESCE(sqlc.narg(amount), amount),
    units = COALESCE(sqlc.narg(units), units),
    class_id = COALESCE(sqlc.narg(class_id), class_id),
    base_id = COALESCE(sqlc.narg(base_id), base_id),
    price = COALESCE(sqlc.narg(price), price),
    is_purchased = COALESCE(sqlc.narg(is_purchased), is_purchased),
    is_final_product = COALESCE(sqlc.narg(is_final_product), is_final_product)
WHERE id = sqlc.arg(id)
RETURNING 
    id,
    name,
    amount,
    units,
    class_id,
    base_id,
    price,
    is_purchased,
    is_final_product;
 
-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

-- name: GetProductByClassID :many
SELECT
    p.id,
    p.name,
    p.amount,
    p.units,
    p.base_id,
    COALESCE(p.price, 0)::double precision AS price,
    p.is_purchased,
    p.is_final_product,
    c.id AS class_id,
    c.name AS class_name,
    c.parent_id AS class_parent_id,
    c.numeric_parameter_ids AS class_num_params,
    c.enum_parameter_ids AS class_enum_params
FROM products p
JOIN classes c ON p.class_id = c.id
WHERE p.class_id = sqlc.arg(class_id)
ORDER BY p.id
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);
