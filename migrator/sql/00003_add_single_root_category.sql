-- +goose Up
WITH root AS (
	SELECT id
	FROM product_categories
	WHERE parent_id IS NULL
	ORDER BY id
	LIMIT 1
)
UPDATE product_categories pc
SET parent_id = (SELECT id FROM root)
WHERE pc.parent_id IS NULL
  AND pc.id <> (SELECT id FROM root)
  AND EXISTS (SELECT 1 FROM root);

CREATE UNIQUE INDEX idx_single_root_category ON product_categories ((1)) WHERE parent_id IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_single_root_category;

