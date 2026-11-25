-- name: InsertProductRangeValue :exec
INSERT INTO products_parameters (product_id, range_parameter, numeric_value)
VALUES (sqlc.arg(product_id), sqlc.arg(range_parameter_id), sqlc.arg(numeric_value_raw)::text::numeric);

-- name: InsertProductEnumValue :exec
INSERT INTO products_parameters (product_id, enum_parameter_id, enum_value_id)
VALUES (sqlc.arg(product_id), sqlc.arg(enum_parameter_id), sqlc.arg(enum_value_id));

-- name: DeleteProductParameterFields :exec
DELETE FROM products_parameters pp
WHERE pp.product_id = sqlc.arg(product_id)
  AND (
    pp.range_parameter IN (SELECT id FROM range_parameters WHERE field_name = ANY(sqlc.arg(field_names)::text[]))
    OR pp.enum_parameter_id IN (SELECT id FROM enum_parameters WHERE field_name = ANY(sqlc.arg(field_names)::text[]))
  );

-- name: DeleteAllProductParameters :exec
DELETE FROM products_parameters WHERE product_id = $1;

-- name: GetProductParameterValues :many
SELECT pp.product_id, rp.field_name, pp.numeric_value::text AS value_raw, 'float'::text AS value_type
FROM products_parameters pp
JOIN range_parameters rp ON rp.id = pp.range_parameter
WHERE pp.product_id = ANY(sqlc.arg(product_ids)::bigint[])
UNION ALL
SELECT pp.product_id, ep.field_name, pv.value_raw::text AS value_raw, ep.enum_type AS value_type
FROM products_parameters pp
JOIN enum_parameters ep ON ep.id = pp.enum_parameter_id
JOIN parameter_values pv ON pv.id = pp.enum_value_id AND pv.enum_parameter_id = pp.enum_parameter_id
WHERE pp.product_id = ANY(sqlc.arg(product_ids)::bigint[]);

-- name: LockProductForUpdate :one
SELECT class_id, base_id FROM products WHERE id = $1 FOR UPDATE;

-- name: LockClassParameterIDs :one
SELECT numeric_parameter_ids, enum_parameter_ids FROM classes WHERE id = $1 FOR SHARE;

-- name: LockRangeParameterWithValue :one
SELECT id FROM range_parameters
WHERE id = sqlc.arg(range_parameter_id)
  AND sqlc.arg(numeric_value_raw)::text::numeric BETWEEN min_val AND max_val
FOR SHARE;
