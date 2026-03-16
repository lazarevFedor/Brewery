-- +goose Up
CREATE TABLE IF NOT EXISTS beer (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    rating NUMERIC(2, 1) DEFAULT 0.0,
    description TEXT,
    abv float4 NOT NULL,
    ibu SMALLINT NOT NULL,
    features TEXT[],
    city_id SMALLINT NOT NULL,
    category_id SMALLINT NOT NULL,
    CONSTRAINT fk_city FOREIGN KEY (city_id) REFERENCES city (id),
    CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES category (id)
);

-- +goose Down
DROP TABLE IF EXISTS beer;