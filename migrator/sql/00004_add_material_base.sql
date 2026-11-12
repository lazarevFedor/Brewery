-- +goose Up
ALTER TABLE materials
    ADD COLUMN base BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE materials
    DROP COLUMN base;
