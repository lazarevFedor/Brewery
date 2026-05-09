-- +goose Up
CREATE TABLE IF NOT EXISTS numeric_parameters (
    id SERIAL PRIMARY KEY,
    min_val INTEGER,
    max_val INTEGER,
    field_name VARCHAR(50) NOT NULL,
    entity_name VARCHAR(50) NOT NULL,
    inheritable BOOLEAN NOT NULL DEFAULT FALSE
);


CREATE TABLE IF NOT EXISTS enum_parameters (
    id SERIAL PRIMARY KEY,
    enum_class_id INTEGER NOT NULL,
    inheritable BOOLEAN NOT NULL DEFAULT FALSE
);


CREATE TABLE IF NOT EXISTS aggregates (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    numeric_parameter_ids INTEGER[],
    enum_parameter_ids INTEGER[]
);


CREATE INDEX idx_field_name ON numeric_parameters(field_name);
CREATE INDEX idx_enum_class_id ON enum_parameters(enum_class_id);
CREATE INDEX idx_name ON aggregates(name);


ALTER TABLE product_categories
    ADD COLUMN numeric_parameter_ids INTEGER[],
    ADD COLUMN enum_parameter_ids INTEGER[];


-- +goose Down
ALTER TABLE product_categories
    DROP COLUMN numeric_parameter_ids,
    DROP COLUMN enum_parameter_ids;

DROP TABLE IF EXISTS aggregates;
DROP TABLE IF EXISTS enum_parameters;
DROP TABLE IF EXISTS numeric_parameters;