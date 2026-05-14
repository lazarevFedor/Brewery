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


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION get_descendants(p_parent_ids INT[])
RETURNS INT[] AS $$
DECLARE
    v_ids INT[];
BEGIN
    IF array_length(coalesce(p_parent_ids, '{}'), 1) IS NULL THEN
        RETURN '{}';
    END IF;

    WITH RECURSIVE children AS (
        SELECT id FROM product_categories WHERE parent_id = ANY(p_parent_ids)
        UNION ALL
        SELECT pc.id FROM product_categories pc JOIN children c ON pc.parent_id = c.id
    )
    SELECT coalesce(array_agg(id), '{}') INTO v_ids FROM children;

    RETURN v_ids;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION add_numeric_parameters_to_categories(
    p_parameter_ids INT[],
    p_category_ids INT[]
) RETURNS INT AS $$
DECLARE
    v_rows_affected INT := 0;
BEGIN
    IF array_length(p_parameter_ids, 1) = 0 OR array_length(p_category_ids, 1) = 0 THEN
        RETURN 0;
    END IF;

    UPDATE product_categories pc
    SET numeric_parameter_ids = (
        SELECT coalesce(array_agg(DISTINCT x ORDER BY x), '{}')
        FROM unnest(coalesce(pc.numeric_parameter_ids, '{}') || p_parameter_ids) AS x
    )
    WHERE pc.id = ANY(p_category_ids);

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION remove_numeric_parameters_from_categories(
    p_parameter_ids INT[],
    p_category_ids INT[]
) RETURNS INT AS $$
DECLARE
    v_rows_affected INT := 0;
BEGIN
    IF array_length(p_parameter_ids, 1) = 0 OR array_length(p_category_ids, 1) = 0 THEN
        RETURN 0;
    END IF;

    UPDATE product_categories pc
    SET numeric_parameter_ids = (
        SELECT coalesce(array_agg(x ORDER BY x), '{}')
        FROM unnest(coalesce(pc.numeric_parameter_ids, '{}')) AS x
        WHERE NOT (x = ANY(p_parameter_ids))
    )
    WHERE pc.id = ANY(p_category_ids);

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION add_enum_parameters_to_categories(
    p_parameter_ids INT[],
    p_category_ids INT[]
) RETURNS INT AS $$
DECLARE
    v_rows_affected INT := 0;
BEGIN
    IF array_length(p_parameter_ids, 1) = 0 OR array_length(p_category_ids, 1) = 0 THEN
        RETURN 0;
    END IF;

    UPDATE product_categories pc
    SET enum_parameter_ids = (
        SELECT coalesce(array_agg(DISTINCT x ORDER BY x), '{}')
        FROM unnest(coalesce(pc.enum_parameter_ids, '{}') || p_parameter_ids) AS x
    )
    WHERE pc.id = ANY(p_category_ids);

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION remove_enum_parameters_from_categories(
    p_parameter_ids INT[],
    p_category_ids INT[]
) RETURNS INT AS $$
DECLARE
    v_rows_affected INT := 0;
BEGIN
    IF array_length(p_parameter_ids, 1) = 0 OR array_length(p_category_ids, 1) = 0 THEN
        RETURN 0;
    END IF;

    UPDATE product_categories pc
    SET enum_parameter_ids = (
        SELECT coalesce(array_agg(x ORDER BY x), '{}')
        FROM unnest(coalesce(pc.enum_parameter_ids, '{}')) AS x
        WHERE NOT (x = ANY(p_parameter_ids))
    )
    WHERE pc.id = ANY(p_category_ids);

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION inherit_parameters_to_children(
    p_parent_id INT
) RETURNS INT AS $$
DECLARE
    v_inheritable_numeric INT[];
    v_inheritable_enum INT[];
    v_descendants INT[];
    v_rows_affected INT := 0;
BEGIN
    SELECT array_agg(DISTINCT nid)
    INTO v_inheritable_numeric
    FROM unnest(coalesce((SELECT numeric_parameter_ids FROM product_categories WHERE id = p_parent_id), '{}')) AS nid
    JOIN numeric_parameters np ON np.id = nid
    WHERE np.inheritable IS TRUE;

    SELECT array_agg(DISTINCT nid)
    INTO v_inheritable_enum
    FROM unnest(coalesce((SELECT enum_parameter_ids FROM product_categories WHERE id = p_parent_id), '{}')) AS nid
    JOIN enum_parameters ep ON ep.id = nid
    WHERE ep.inheritable IS TRUE;

    v_descendants := get_descendants(ARRAY[p_parent_id]);

    IF coalesce(array_length(v_inheritable_numeric, 1), 0) > 0 AND coalesce(array_length(v_descendants, 1), 0) > 0 THEN
        v_rows_affected := v_rows_affected + add_numeric_parameters_to_categories(v_inheritable_numeric, v_descendants);
    END IF;

    IF coalesce(array_length(v_inheritable_enum, 1), 0) > 0 AND coalesce(array_length(v_descendants, 1), 0) > 0 THEN
        v_rows_affected := v_rows_affected + add_enum_parameters_to_categories(v_inheritable_enum, v_descendants);
    END IF;

    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION remove_numeric_parameters_from_descendants(
    p_parameter_ids INT[],
    p_parent_ids INT[]
) RETURNS INT AS $$
DECLARE
    v_descendants INT[];
BEGIN
    IF array_length(coalesce(p_parameter_ids, '{}'), 1) = 0 OR array_length(coalesce(p_parent_ids, '{}'), 1) = 0 THEN
        RETURN 0;
    END IF;

    v_descendants := get_descendants(p_parent_ids);
    IF array_length(coalesce(v_descendants, '{}'), 1) = 0 THEN
        RETURN 0;
    END IF;

    RETURN remove_numeric_parameters_from_categories(p_parameter_ids, v_descendants);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION remove_enum_parameters_from_descendants(
    p_parameter_ids INT[],
    p_parent_ids INT[]
) RETURNS INT AS $$
DECLARE
    v_descendants INT[];
BEGIN
    IF array_length(coalesce(p_parameter_ids, '{}'), 1) = 0 OR array_length(coalesce(p_parent_ids, '{}'), 1) = 0 THEN
        RETURN 0;
    END IF;

    v_descendants := get_descendants(p_parent_ids);
    IF array_length(coalesce(v_descendants, '{}'), 1) = 0 THEN
        RETURN 0;
    END IF;

    RETURN remove_enum_parameters_from_categories(p_parameter_ids, v_descendants);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION cleanup_numeric_parameter_from_categories()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE product_categories
    SET numeric_parameter_ids = array_remove(numeric_parameter_ids, OLD.id)
    WHERE numeric_parameter_ids @> ARRAY[OLD.id];

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE TRIGGER trg_before_delete_numeric_parameter
BEFORE DELETE ON numeric_parameters
FOR EACH ROW
EXECUTE FUNCTION cleanup_numeric_parameter_from_categories();
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION cleanup_enum_parameter_from_categories()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE product_categories
    SET enum_parameter_ids = array_remove(enum_parameter_ids, OLD.id)
    WHERE enum_parameter_ids @> ARRAY[OLD.id];

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE TRIGGER trg_before_delete_enum_parameter
BEFORE DELETE ON enum_parameters
FOR EACH ROW
EXECUTE FUNCTION cleanup_enum_parameter_from_categories();
-- +goose StatementEnd


-- +goose Down
DROP TRIGGER IF EXISTS trg_before_delete_enum_parameter ON enum_parameters;
DROP TRIGGER IF EXISTS trg_before_delete_numeric_parameter ON numeric_parameters;
DROP FUNCTION IF EXISTS cleanup_enum_parameter_from_categories();
DROP FUNCTION IF EXISTS cleanup_numeric_parameter_from_categories();
DROP FUNCTION IF EXISTS remove_enum_parameters_from_descendants(INT[], INT[]);
DROP FUNCTION IF EXISTS remove_numeric_parameters_from_descendants(INT[], INT[]);
DROP FUNCTION IF EXISTS inherit_parameters_to_children(INT);
DROP FUNCTION IF EXISTS get_descendants(INT[]);
DROP FUNCTION IF EXISTS remove_enum_parameters_from_categories(INT[], INT[]);
DROP FUNCTION IF EXISTS add_enum_parameters_to_categories(INT[], INT[]);
DROP FUNCTION IF EXISTS remove_numeric_parameters_from_categories(INT[], INT[]);
DROP FUNCTION IF EXISTS add_numeric_parameters_to_categories(INT[], INT[]);

ALTER TABLE product_categories
    DROP COLUMN numeric_parameter_ids,
    DROP COLUMN enum_parameter_ids;

DROP TABLE IF EXISTS aggregates;
DROP TABLE IF EXISTS enum_parameters;
DROP TABLE IF EXISTS numeric_parameters;