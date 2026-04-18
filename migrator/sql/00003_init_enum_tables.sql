-- +goose Up

CREATE TABLE enum_classes (
    id INT PRIMARY KEY,
    enum_type VARCHAR(100)
    entity_name VARCHAR(50),
    field_name VARCHAR(50),

);

CREATE TABLE enum_values (
    id INT PRIMARY KEY,
    enum_class_id INT NOT NULL,

    value_raw VARCHAR(255) NOT NULL,   -- само значение
    value_type VARCHAR(20) NOT NULL,   -- 'int', 'float', 'string', 'picture'

    position INT

    FOREIGN KEY (enum_class_id) REFERENCES enum_classes(id)
);

-- +goose Down
