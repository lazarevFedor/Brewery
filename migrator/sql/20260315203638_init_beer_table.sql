-- +goose Up
CREATE TABLE IF NOT EXISTS beer (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    rating NUMERIC(2, 1) DEFAULT 0.0,
    description TEXT,
    abv float4 NOT NULL,
    ibu SMALLINT NOT NULL,
    features TEXT[],
    city_id SMALLINT NOT NULL REFERENCES city.id,
    category_id SMALLINT NOT NULL REFERENCES category.id
);

-- +goose Down
DROP TABLE IF EXISTS beer;