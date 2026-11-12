-- name: InsertChange :one
INSERT INTO change_history (
    product_id,
    changes
) VALUES (
    sqlc.arg(product_id),
    sqlc.arg(changes)
) RETURNING
    id,
    product_id,
    changes,
    created_at;

-- name: GetHistoryRecordByID :many
SELECT 
    id,
    product_id,
    changes,
    created_at
FROM change_history
WHERE product_id = $1;