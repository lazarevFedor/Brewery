-- +goose Up
-- ==========================================
--                TABLES
-- ==========================================
CREATE TABLE IF NOT EXISTS enum_parameters (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    enum_type TEXT NOT NULL, 
    field_name TEXT NOT NULL,
    unit TEXT,
    inheritable BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS parameter_values (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    enum_parameter_id BIGINT REFERENCES enum_parameters(id) ON DELETE CASCADE,
    value_raw varchar(256) NOT NULL,
    position INT,
    CONSTRAINT chk_enum_values_position_positive CHECK (position > 0),
    CONSTRAINT uq_parameter_values_id_parameter UNIQUE (id, enum_parameter_id)
);

CREATE TABLE IF NOT EXISTS range_parameters (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    min_val NUMERIC(8, 3),
    max_val NUMERIC(8, 3),
    unit TEXT NOT NULL,
    field_name TEXT NOT NULL,
    inheritable BOOLEAN NOT NULL,
    CONSTRAINT chk_min_lower_max CHECK (min_val <= max_val)
);


CREATE TABLE IF NOT EXISTS products_parameters (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    enum_value_id BIGINT,
    range_parameter BIGINT REFERENCES range_parameters(id) ON DELETE RESTRICT,
    numeric_value NUMERIC(8, 3),
    enum_parameter_id BIGINT REFERENCES enum_parameters(id) ON DELETE RESTRICT,
    CONSTRAINT fk_products_parameters_enum_value
        FOREIGN KEY (enum_value_id, enum_parameter_id)
        REFERENCES parameter_values(id, enum_parameter_id) ON DELETE RESTRICT,
    CONSTRAINT chk_products_parameters_value_kind CHECK (
        (
            range_parameter IS NOT NULL AND numeric_value IS NOT NULL
            AND enum_parameter_id IS NULL AND enum_value_id IS NULL
        ) OR (
            range_parameter IS NULL AND numeric_value IS NULL
            AND enum_parameter_id IS NOT NULL AND enum_value_id IS NOT NULL
        )
    )
);

CREATE TABLE IF NOT EXISTS aggregates (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL, 
    description TEXT NOT NULL DEFAULT '',
    numeric_parameter_ids BIGINT[],
    enum_parameter_ids BIGINT[]
);

-- ==========================================
--                CONSTRAINTS
-- ==========================================

-- Индекс для защиты от повторяющихся параметров
CREATE UNIQUE INDEX idx_product_parameters_unique_range
ON products_parameters (product_id, range_parameter)
WHERE range_parameter IS NOT NULL;

-- Индекс для защиты от повторяющихся параметров
CREATE UNIQUE INDEX idx_product_parameters_unique_enum
ON products_parameters (product_id, enum_parameter_id)
WHERE enum_parameter_id IS NOT NULL;


-- ==========================================
--                INDEXES
-- ==========================================

-- Индекс для поиска параметров по текстовому имени поля
CREATE INDEX IF NOT EXISTS idx_enum_params_field_name
ON enum_parameters(field_name);

-- Индекс для поиска значений по id параметра
CREATE INDEX IF NOT EXISTS parameters_values_enum_parameters_id 
ON parameter_values(enum_parameter_id);


-- Индекс для поиска параметров по id товара
CREATE INDEX IF NOT EXISTS idx_products_parameters_product_id 
ON products_parameters(product_id);



-- +goose Down
DROP TABLE IF EXISTS products_parameters;
DROP TABLE IF EXISTS aggregates;
DROP TABLE IF EXISTS range_parameters;
DROP TABLE IF EXISTS parameter_values;
DROP TABLE IF EXISTS enum_parameters;
