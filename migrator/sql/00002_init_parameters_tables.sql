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

-- +goose StatementBegin
CREATE FUNCTION class_parameter_name_counts(range_ids BIGINT[], enum_ids BIGINT[])
RETURNS TABLE (field_name TEXT, definition_count BIGINT)
LANGUAGE sql STABLE AS $$
    SELECT names.field_name, count(*)
    FROM (
        SELECT rp.field_name FROM range_parameters rp WHERE rp.id = ANY(range_ids)
        UNION ALL
        SELECT ep.field_name FROM enum_parameters ep WHERE ep.id = ANY(enum_ids)
    ) AS names
    GROUP BY names.field_name;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION check_class_parameter_names()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    old_range_ids BIGINT[];
    old_enum_ids BIGINT[];
    duplicate_name TEXT;
BEGIN
    IF TG_OP = 'UPDATE' THEN
        old_range_ids := OLD.numeric_parameter_ids;
        old_enum_ids := OLD.enum_parameter_ids;
    END IF;

    SELECT proposed.field_name INTO duplicate_name
    FROM class_parameter_name_counts(NEW.numeric_parameter_ids, NEW.enum_parameter_ids) proposed
    LEFT JOIN class_parameter_name_counts(old_range_ids, old_enum_ids) previous USING (field_name)
    WHERE proposed.definition_count > 1
      AND proposed.definition_count > COALESCE(previous.definition_count, 0)
    ORDER BY proposed.field_name
    LIMIT 1;

    IF FOUND THEN
        RAISE unique_violation USING CONSTRAINT = 'class_parameter_field_name_unique',
            MESSAGE = format('Parameter field %L is already used by class %s', duplicate_name, NEW.id);
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_check_class_parameter_names
BEFORE INSERT OR UPDATE OF numeric_parameter_ids, enum_parameter_ids ON classes
FOR EACH ROW EXECUTE FUNCTION check_class_parameter_names();

-- +goose StatementBegin
CREATE FUNCTION check_class_parameter_rename()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    conflicting_class BIGINT;
BEGIN
    SELECT c.id INTO conflicting_class
    FROM classes c
    WHERE CASE WHEN TG_TABLE_NAME = 'range_parameters' THEN NEW.id = ANY(c.numeric_parameter_ids)
               ELSE NEW.id = ANY(c.enum_parameter_ids) END
      AND EXISTS (
          SELECT 1
          FROM class_parameter_name_counts(
              CASE WHEN TG_TABLE_NAME = 'range_parameters' THEN array_remove(c.numeric_parameter_ids, NEW.id)
                   ELSE c.numeric_parameter_ids END,
              CASE WHEN TG_TABLE_NAME = 'enum_parameters' THEN array_remove(c.enum_parameter_ids, NEW.id)
                   ELSE c.enum_parameter_ids END
          ) other WHERE other.field_name = NEW.field_name
      )
    ORDER BY c.id
    LIMIT 1;

    IF FOUND THEN
        RAISE unique_violation USING CONSTRAINT = 'class_parameter_field_name_unique',
            MESSAGE = format('Parameter field %L is already used by class %s', NEW.field_name, conflicting_class);
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_check_enum_parameter_rename
BEFORE UPDATE OF field_name ON enum_parameters
FOR EACH ROW WHEN (OLD.field_name IS DISTINCT FROM NEW.field_name)
EXECUTE FUNCTION check_class_parameter_rename();

CREATE TRIGGER trg_check_range_parameter_rename
BEFORE UPDATE OF field_name ON range_parameters
FOR EACH ROW WHEN (OLD.field_name IS DISTINCT FROM NEW.field_name)
EXECUTE FUNCTION check_class_parameter_rename();

-- +goose Down
DROP TRIGGER trg_check_range_parameter_rename ON range_parameters;
DROP TRIGGER trg_check_enum_parameter_rename ON enum_parameters;
DROP TRIGGER trg_check_class_parameter_names ON classes;
DROP FUNCTION check_class_parameter_rename();
DROP FUNCTION check_class_parameter_names();
DROP FUNCTION class_parameter_name_counts(BIGINT[], BIGINT[]);

DROP TABLE IF EXISTS products_parameters;
DROP TABLE IF EXISTS aggregates;
DROP TABLE IF EXISTS range_parameters;
DROP TABLE IF EXISTS parameter_values;
DROP TABLE IF EXISTS enum_parameters;
