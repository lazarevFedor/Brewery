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

-- name: RenameClass :one
UPDATE classes
SET name = sqlc.arg(name)
WHERE id = sqlc.arg(id)
RETURNING 
    id, 
    name, 
    parent_id,
    numeric_parameter_ids,
    enum_parameter_ids;

-- name: UpdateParentClass :one
UPDATE classes
SET parent_id = sqlc.arg(parent_id)
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