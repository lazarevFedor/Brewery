-- +goose Up
CREATE TABLE IF NOT EXISTS city(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    country_id SMALLINT,
    CONSTRAINT fk_country FOREIGN KEY (country_id) REFERENCES country (id)
);

-- +goose Down
DROP TABLE IF EXISTS city;