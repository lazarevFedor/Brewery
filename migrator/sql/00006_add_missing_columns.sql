-- +goose Up

-- Колонки добавлены в 00003, но могли отсутствовать если таблица создавалась до их появления
ALTER TABLE enum_classes
    ADD COLUMN IF NOT EXISTS unit VARCHAR(16),
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

-- Колонки добавлены в 00004, но могли отсутствовать если 00004 не применялась поверх старой схемы
ALTER TABLE beers
    ADD COLUMN IF NOT EXISTS review_amount INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS review_rating_sum INT DEFAULT 0;

-- +goose Down
ALTER TABLE enum_classes
    DROP COLUMN IF EXISTS unit,
    DROP COLUMN IF EXISTS is_active;

ALTER TABLE beers
    DROP COLUMN IF EXISTS review_amount,
    DROP COLUMN IF EXISTS review_rating_sum;
