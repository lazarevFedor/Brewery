-- name: ExcludeEnumParameters :exec
DELETE FROM products_parameters pp
USING parameter_values p
WHERE p.id = pp.enum_value_id 
  AND pp.product_id = sqlc.arg(product_id) 
  AND p.enum_parameter_id = ANY(sqlc.arg(enum_parameter_ids)::bigint[]);


-- name: ExcludeRangeParameters :exec
DELETE FROM products_parameters
WHERE product_id = sqlc.arg(product_id)
  AND range_parameter = ANY(sqlc.arg(range_parameter_ids)::bigint[]);