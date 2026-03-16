-- +goose Up
CREATE TABLE IF NOT EXISTS review(
    id SERIAL PRIMARY KEY,
    body TEXT,
    rating NUMERIC(2, 1) NOT NULL,
    beer_id INT,
    CONSTRAINT fk_review_beer FOREIGN KEY (beer_id) REFERENCES beer (id)
);

-- +goose Down
DROP TABLE IF EXISTS review;