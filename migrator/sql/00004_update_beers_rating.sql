-- +goose Up
ALTER TABLE beers
    DROP COLUMN rating,
    ADD review_amount INT DEFAULT 0,
    ADD review_rating_sum INT DEFAULT 0;

ALTER TABLE reviews
    ALTER COLUMN rating TYPE INT,
    ALTER COLUMN rating SET NOT NULL,
    ADD CONSTRAINT check_rating_range CHECK (rating >= 1);


-- +goose Down
ALTER TABLE beers 
    ADD COLUMN rating NUMERIC(2, 1) DEFAULT 0.0,
    DROP COLUMN review_amount,
    DROP COLUMN review_rating_sum;

ALTER TABLE reviews 
    ALTER COLUMN rating TYPE NUMERIC(2, 1),
    ALTER COLUMN rating SET NOT NULL,
    DROP CONSTRAINT check_rating_range; 
