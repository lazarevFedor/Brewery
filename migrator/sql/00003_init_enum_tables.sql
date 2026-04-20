-- +goose Up

CREATE TABLE IF NOT EXISTS enum_classes (
    id INT PRIMARY KEY,
    enum_type VARCHAR(100) NOT NULL,
    entity_name VARCHAR(50) NOT NULL,
    field_name VARCHAR(50) NOT NULL,
    unit VARCHAR(16),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    CHECK (
        (enum_type IN ('int', 'float') AND (unit IS NULL OR BTRIM(unit) <> '')) OR
        (enum_type NOT IN ('int', 'float') AND unit IS NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_enum_classes_numeric_field_type_unit
    ON enum_classes (entity_name, field_name, enum_type, COALESCE(unit, ''))
    WHERE enum_type IN ('int', 'float');

CREATE UNIQUE INDEX IF NOT EXISTS uq_enum_classes_single_active
    ON enum_classes (entity_name, field_name)
    WHERE is_active;

CREATE TABLE IF NOT EXISTS enum_values (
    id INT PRIMARY KEY,
    enum_class_id INT NOT NULL,

    value_raw VARCHAR(255) NOT NULL,   -- само значение
    value_type VARCHAR(20) NOT NULL,   -- 'int', 'float', 'string', 'picture'

    position INT NOT NULL,
    FOREIGN KEY (enum_class_id) REFERENCES enum_classes(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS enum_values;
DROP TABLE IF EXISTS enum_classes;
