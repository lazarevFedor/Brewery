-- name: InsertEquipment :one
INSERT INTO work_centers (equipment_name, work_center_name)
VALUES ($1, $2)
RETURNING equipment_name, work_center_name;

-- name: ListEquipment :many
SELECT equipment_name, work_center_name
FROM work_centers
WHERE (sqlc.narg(name_filter)::text IS NULL OR equipment_name = sqlc.narg(name_filter))
  AND (sqlc.narg(center_filter)::text IS NULL OR work_center_name = sqlc.narg(center_filter))
ORDER BY equipment_name, work_center_name
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: GetEquipmentByKey :one
SELECT equipment_name, work_center_name
FROM work_centers
WHERE equipment_name = $1 AND work_center_name = $2;

-- name: UpdateEquipment :one
UPDATE work_centers
SET equipment_name = COALESCE(sqlc.narg(new_name)::text, equipment_name),
    work_center_name = COALESCE(sqlc.narg(new_center)::text, work_center_name)
WHERE equipment_name = sqlc.arg(equipment_name) AND work_center_name = sqlc.arg(work_center_name)
RETURNING equipment_name, work_center_name;

-- name: DeleteEquipment :execrows
DELETE FROM work_centers
WHERE equipment_name = $1 AND work_center_name = $2;
