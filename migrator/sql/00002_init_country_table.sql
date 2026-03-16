-- +goose Up
CREATE TABLE IF NOT EXISTS country(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS country;
