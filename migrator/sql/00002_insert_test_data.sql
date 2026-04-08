-- +goose Up
INSERT INTO product_categories (name)
VALUES ('test_category');

INSERT INTO countries (name)
VALUES ('test_country');

INSERT INTO cities (name, country_id)
VALUES ('test_city', 1);

-- +goose Down
SELECT 'down SQL query';
