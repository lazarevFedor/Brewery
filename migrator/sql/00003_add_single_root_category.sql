-- +goose Up
CREATE UNIQUE INDEX idx_single_root_category ON product_categories(1) WHERE parent_id IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_single_root_category;

