-- +goose Up
CREATE TABLE IF NOT EXISTS category(
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE,
    parent_id SMALLINT,
    CONSTRAINT fk_category_parent FOREIGN KEY (parent_id) REFERENCES category (id)
);

-- +goose Down
DROP TABLE IF EXISTS category;