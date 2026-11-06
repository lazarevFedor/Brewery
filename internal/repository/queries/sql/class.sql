-- name: GetClasses :many
SELECT
    id,
    name,
    parent_id,
    numeric_parameter_ids,
    enum_parameter_ids
FROM classes;

-- name: InsertClass :one
INSERT INTO classes(
    name,
    parent_id
) VALUES (
    sqlc.arg(name),
    sqlc.narg(parent_id)
) RETURNING 
    id, 
    name, 
    parent_id;

-- name: GetClassByID :one
SELECT
    id, 
    name, 
    parent_id,
    numeric_parameter_ids,
    enum_parameter_ids
FROM classes
WHERE id = $1;

-- name: UpdateClass :one
UPDATE classes
SET 
    name = COALESCE(sqlc.narg(name), name),
    parent_id = COALESCE(sqlc.narg(parent_id), parent_id),
    numeric_parameter_ids = COALESCE(sqlc.narg(numeric_parameter_ids), numeric_parameter_ids),
    enum_parameter_ids = COALESCE(sqlc.narg(enum_parameter_ids), enum_parameter_ids)
WHERE id = sqlc.arg(id)
RETURNING 
    id, 
    name, 
    parent_id,
    numeric_parameter_ids,
    enum_parameter_ids;

-- name: DeleteClass :exec
DELETE FROM classes
WHERE id = $1;

-- name: GetClassByName :one
SELECT
    id,
    name,
    parent_id,
    numeric_parameter_ids,
    enum_parameter_ids
FROM classes
WHERE name = $1;