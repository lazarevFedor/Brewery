-- +goose Up
INSERT INTO product_categories (name)
VALUES ('test_category');


-- +goose Down
SELECT 'down SQL query';
