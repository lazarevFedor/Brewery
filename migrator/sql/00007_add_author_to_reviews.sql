-- +goose Up
ALTER TABLE reviews
    ADD COLUMN IF NOT EXISTS author VARCHAR(100);

-- +goose Down
ALTER TABLE reviews
    DROP COLUMN IF EXISTS author;
