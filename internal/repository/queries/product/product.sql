-- name: :one/many/exec



-- name: InsertProduct :one
INSERT INTO products(
    name,
    amount,
    units,
    base_id
) VALUES (
    $1, $2, $3, $4
) RETURNING 
    id,
    name,
    amount,
    units,
    base_id;

-- name: GetProductByID :one
SELECT id, name, 
        amount, units,
        c.id,
        c.name,
        c.parent_id,
        c.numeric_parameters_ids,
        c.enum_parameter_ids,
        base_id
JOIN classes c ON p.class_id = c.id
FROM products p
WHERE id = ($1);
 

-- name: GetProducts :many
SELECT id, name, 
        amount, units,
        class_id, base_id
FROM products;
 
-- name: UpdateProduct :one
UPDATE products
SET
    name = COALESCE(sqlc.arg(name), name),
    amount = COALESCE(sqlc.arg(amount), amount),
    units = COALESCE(sqlc.arg(units), units),
    class_id = COALESCE(sqlc.arg(class_id), class_id),
    base_id = COALESCE(sqlc.arg(base_id), base_id)
WHERE id = sqlc.arg(id)
FROM products
RETURNING 
    id,
    name,
    amount,
    units,
    base_id;
 
-- name: DeleteProduct :exec
DELETE products
WHERE id = @id;

-- name: GetProductByClassID :many
SELECT id, name, 
        amount, units,
        class_id, base_id
WHERE class_id = @class_id;