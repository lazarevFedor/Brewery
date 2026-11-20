-- name: InsertOperationRangeValue :exec
INSERT INTO operations_parameters (operation_id, range_parameter, numeric_value)
VALUES (sqlc.arg(operation_id), sqlc.arg(range_parameter_id), sqlc.arg(numeric_value_raw)::text::numeric);

-- name: InsertOperationEnumValue :exec
INSERT INTO operations_parameters (operation_id, enum_parameter_id, enum_value_id)
VALUES (sqlc.arg(operation_id), sqlc.arg(enum_parameter_id), sqlc.arg(enum_value_id));

-- name: DeleteOperationParameterFields :exec
DELETE FROM operations_parameters op
WHERE op.operation_id = sqlc.arg(operation_id)
  AND (
    op.range_parameter IN (SELECT id FROM range_parameters WHERE field_name = ANY(sqlc.arg(field_names)::text[]))
    OR op.enum_parameter_id IN (SELECT id FROM enum_parameters WHERE field_name = ANY(sqlc.arg(field_names)::text[]))
  );

-- name: DeleteAllOperationParameters :exec
DELETE FROM operations_parameters WHERE operation_id = $1;

-- name: GetOperationParameterValues :many
SELECT op.operation_id, rp.field_name, op.numeric_value::text AS value_raw, 'float'::text AS value_type
FROM operations_parameters op
JOIN range_parameters rp ON rp.id = op.range_parameter
WHERE op.operation_id = ANY(sqlc.arg(operation_ids)::bigint[])
UNION ALL
SELECT op.operation_id, ep.field_name, pv.value_raw::text AS value_raw, ep.enum_type AS value_type
FROM operations_parameters op
JOIN enum_parameters ep ON ep.id = op.enum_parameter_id
JOIN parameter_values pv ON pv.id = op.enum_value_id AND pv.enum_parameter_id = op.enum_parameter_id
WHERE op.operation_id = ANY(sqlc.arg(operation_ids)::bigint[]);
