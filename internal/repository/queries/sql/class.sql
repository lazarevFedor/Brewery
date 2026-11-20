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
    parent_id = COALESCE(sqlc.narg(parent_id), parent_id)
WHERE id = sqlc.arg(id)
RETURNING 
    id, 
    name, 
    parent_id,
    numeric_parameter_ids,
    enum_parameter_ids;

-- name: DeleteClass :execrows
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

-- name: ClassMoveCreatesCycle :one
WITH RECURSIVE ancestors(id, parent_id) AS (
    SELECT c.id, c.parent_id FROM classes c WHERE c.id = sqlc.arg(class_id)
    UNION
    SELECT c.id, c.parent_id FROM classes c JOIN ancestors a ON c.id = a.parent_id
)
SELECT (COALESCE(bool_or(id = sqlc.arg(ancestor_id)::bigint), false)
        OR (count(*) > 0 AND count(*) FILTER (WHERE parent_id IS NULL) = 0))::boolean
FROM ancestors;

-- name: InheritClassBranchParameters :exec
WITH RECURSIVE branch(id) AS (
    SELECT child.id FROM classes child WHERE child.id = sqlc.arg(class_id)
    UNION
    SELECT child.id FROM classes child JOIN branch b ON child.parent_id = b.id
), incoming AS (
    SELECT ARRAY(
        SELECT rp.id FROM range_parameters rp
        WHERE rp.id = ANY(parent.numeric_parameter_ids) AND rp.inheritable
    ) AS range_ids, ARRAY(
        SELECT ep.id FROM enum_parameters ep
        WHERE ep.id = ANY(parent.enum_parameter_ids) AND ep.inheritable
    ) AS enum_ids
    FROM classes parent
    WHERE parent.id = sqlc.arg(parent_id)::bigint
)
UPDATE classes AS target
SET numeric_parameter_ids = CASE
        WHEN incoming.range_ids <@ COALESCE(target.numeric_parameter_ids, '{}') THEN target.numeric_parameter_ids
        ELSE ARRAY(SELECT unnest(target.numeric_parameter_ids) UNION SELECT unnest(incoming.range_ids) ORDER BY 1)
    END,
    enum_parameter_ids = CASE
        WHEN incoming.enum_ids <@ COALESCE(target.enum_parameter_ids, '{}') THEN target.enum_parameter_ids
        ELSE ARRAY(SELECT unnest(target.enum_parameter_ids) UNION SELECT unnest(incoming.enum_ids) ORDER BY 1)
    END
FROM branch, incoming
WHERE target.id = branch.id
  AND (NOT incoming.range_ids <@ COALESCE(target.numeric_parameter_ids, '{}')
       OR NOT incoming.enum_ids <@ COALESCE(target.enum_parameter_ids, '{}'));
