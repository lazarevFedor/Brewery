-- +goose Up

CREATE TABLE IF NOT EXISTS enum_classes (
    id INT PRIMARY KEY,
    enum_type VARCHAR(100) NOT NULL,
    entity_name VARCHAR(50) NOT NULL,
    field_name VARCHAR(50) NOT NULL,
    UNIQUE (entity_name, field_name)
);

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
