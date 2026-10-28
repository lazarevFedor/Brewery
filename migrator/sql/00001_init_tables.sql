-- +goose Up
CREATE TABLE IF NOT EXISTS product_categories(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    parent_id INT REFERENCES product_categories (id)
);


CREATE TABLE IF NOT EXISTS countries(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT countries_name_unique UNIQUE (name)
);


CREATE TABLE IF NOT EXISTS cities(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    country_id SMALLINT,
    CONSTRAINT fk_country FOREIGN KEY (country_id) REFERENCES countries (id),
    CONSTRAINT cities_name_country_id_unique UNIQUE (name, country_id)
);


CREATE TABLE IF NOT EXISTS beers (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    abv float4 NOT NULL,
    ibu SMALLINT NOT NULL,
    amount INT NOT NULL,
    units text NOT NULL,
    city_id SMALLINT NOT NULL,
    category_id INT NOT NULL,
    CONSTRAINT fk_city FOREIGN KEY (city_id) REFERENCES cities (id),
    CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES product_categories (id)
);
CREATE INDEX idx_beer_name ON beers(name);


CREATE TABLE IF NOT EXISTS features (
    id SERIAL PRIMARY KEY,
    name TEXT, 
    CONSTRAINT features_name_unique UNIQUE (name)
);


CREATE TABLE IF NOT EXISTS beer_features (
    beer_id INT,
    feature_id INT,
    PRIMARY KEY (beer_id, feature_id)
);



-- +goose Down
DROP TABLE IF EXISTS beer_features;
DROP TABLE IF EXISTS beers;
DROP TABLE IF EXISTS cities;
DROP TABLE IF EXISTS features;
DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS countries;