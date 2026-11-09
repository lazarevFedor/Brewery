-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_enum_values_reorder_before()
RETURNS trigger AS
$$
DECLARE
    max_pos integer;
BEGIN
    -- Internal position shifts executed by this trigger must not trigger reordering again.
    IF pg_trigger_depth() > 1 THEN
        RETURN NEW;
    END IF;

    IF TG_OP = 'INSERT' THEN
        PERFORM 1 FROM parameter_values WHERE enum_parameter_id = NEW.enum_parameter_id FOR UPDATE;

        SELECT COUNT(*) INTO max_pos
        FROM parameter_values
        WHERE enum_parameter_id = NEW.enum_parameter_id;

        IF NEW.position IS NULL OR NEW.position < 1 THEN
            NEW.position := max_pos + 1;
        ELSIF NEW.position > max_pos + 1 THEN
            NEW.position := max_pos + 1;
        END IF;

        UPDATE parameter_values
        SET position = position + 1
        WHERE enum_parameter_id = NEW.enum_parameter_id
          AND position >= NEW.position;

        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        IF NEW.position IS NULL THEN
            NEW.position := OLD.position;
        END IF;

        IF NEW.enum_parameter_id = OLD.enum_parameter_id THEN
            PERFORM 1 FROM parameter_values WHERE enum_parameter_id = NEW.enum_parameter_id FOR UPDATE;

            SELECT COUNT(*) INTO max_pos
            FROM parameter_values
            WHERE enum_parameter_id = NEW.enum_parameter_id
              AND id <> OLD.id;

            IF NEW.position < 1 THEN
                NEW.position := 1;
            ELSIF NEW.position > max_pos + 1 THEN
                NEW.position := max_pos + 1;
            END IF;

            IF NEW.position < OLD.position THEN
                UPDATE parameter_values
                SET position = position + 1
                WHERE enum_parameter_id = OLD.enum_parameter_id
                  AND id <> OLD.id
                  AND position >= NEW.position
                  AND position < OLD.position;
            ELSIF NEW.position > OLD.position THEN
                UPDATE parameter_values
                SET position = position - 1
                WHERE enum_parameter_id = OLD.enum_parameter_id
                  AND id <> OLD.id
                  AND position > OLD.position
                  AND position <= NEW.position;
            END IF;

            RETURN NEW;
        END IF;

        -- Move between classes: compact old class and insert in new class position.
        PERFORM 1 FROM parameter_values WHERE enum_parameter_id IN (OLD.enum_parameter_id, NEW.enum_parameter_id) FOR UPDATE;

        UPDATE parameter_values
        SET position = position - 1
        WHERE enum_parameter_id = OLD.enum_parameter_id
          AND position > OLD.position;

        SELECT COUNT(*) INTO max_pos
        FROM parameter_values
        WHERE enum_parameter_id = NEW.enum_parameter_id;

        IF NEW.position < 1 THEN
            NEW.position := 1;
        ELSIF NEW.position > max_pos + 1 THEN
            NEW.position := max_pos + 1;
        END IF;

        UPDATE parameter_values
        SET position = position + 1
        WHERE enum_parameter_id = NEW.enum_parameter_id
          AND position >= NEW.position;

        RETURN NEW;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_enum_values_reorder_after_delete()
RETURNS trigger AS
$$
BEGIN
    IF pg_trigger_depth() > 1 THEN
        RETURN OLD;
    END IF;

    UPDATE parameter_values
    SET position = position - 1
    WHERE enum_parameter_id = OLD.enum_parameter_id
      AND position > OLD.position;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_enum_values_reorder_before
BEFORE INSERT OR UPDATE ON parameter_values
FOR EACH ROW
EXECUTE FUNCTION fn_enum_values_reorder_before();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_enum_values_reorder_after_delete
AFTER DELETE ON parameter_values
FOR EACH ROW
EXECUTE FUNCTION fn_enum_values_reorder_after_delete();
-- +goose StatementEnd



-- +goose StatementBegin
CREATE OR REPLACE FUNCTION get_descendants(p_parent_ids BIGINT[])
RETURNS BIGINT[] AS $$
DECLARE
    v_ids BIGINT[];
BEGIN
    IF array_length(coalesce(p_parent_ids, '{}'), 1) IS NULL THEN
        RETURN '{}';
    END IF;

    WITH RECURSIVE children AS (
        SELECT id FROM classes WHERE parent_id = ANY(p_parent_ids)
        UNION ALL
        SELECT c.id FROM classes c JOIN children ch ON c.parent_id = ch.id
    )
    SELECT coalesce(array_agg(id), '{}') INTO v_ids FROM children;

    RETURN v_ids;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION add_numeric_parameters_to_categories(
    p_parameter_ids BIGINT[],
    p_category_ids BIGINT[]
) RETURNS INT AS $$
DECLARE
    v_rows_affected INT := 0;
BEGIN
    IF array_length(p_parameter_ids, 1) = 0 OR array_length(p_category_ids, 1) = 0 THEN
        RETURN 0;
    END IF;

    UPDATE classes c
    SET numeric_parameter_ids = (
        SELECT coalesce(array_agg(DISTINCT x ORDER BY x), '{}')
        FROM unnest(coalesce(c.numeric_parameter_ids, '{}') || p_parameter_ids) AS x
    )
    WHERE c.id = ANY(p_category_ids);

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION remove_numeric_parameters_from_categories(
    p_parameter_ids BIGINT[],
    p_category_ids BIGINT[]
) RETURNS INT AS $$
DECLARE
    v_rows_affected INT := 0;
BEGIN
    IF array_length(p_parameter_ids, 1) = 0 OR array_length(p_category_ids, 1) = 0 THEN
        RETURN 0;
    END IF;

    UPDATE classes c
    SET numeric_parameter_ids = (
        SELECT coalesce(array_agg(x ORDER BY x), '{}')
        FROM unnest(coalesce(c.numeric_parameter_ids, '{}')) AS x
        WHERE NOT (x = ANY(p_parameter_ids))
    )
    WHERE c.id = ANY(p_category_ids);

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION add_enum_parameters_to_categories(
    p_parameter_ids BIGINT[],
    p_category_ids BIGINT[]
) RETURNS INT AS $$
DECLARE
    v_rows_affected INT := 0;
BEGIN
    IF array_length(p_parameter_ids, 1) = 0 OR array_length(p_category_ids, 1) = 0 THEN
        RETURN 0;
    END IF;

    UPDATE classes c
    SET enum_parameter_ids = (
        SELECT coalesce(array_agg(DISTINCT x ORDER BY x), '{}')
        FROM unnest(coalesce(c.enum_parameter_ids, '{}') || p_parameter_ids) AS x
    )
    WHERE c.id = ANY(p_category_ids);

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION remove_enum_parameters_from_categories(
    p_parameter_ids BIGINT[],
    p_category_ids BIGINT[]
) RETURNS INT AS $$
DECLARE
    v_rows_affected INT := 0;
BEGIN
    IF array_length(p_parameter_ids, 1) = 0 OR array_length(p_category_ids, 1) = 0 THEN
        RETURN 0;
    END IF;

    UPDATE classes c
    SET enum_parameter_ids = (
        SELECT coalesce(array_agg(x ORDER BY x), '{}')
        FROM unnest(coalesce(c.enum_parameter_ids, '{}')) AS x
        WHERE NOT (x = ANY(p_parameter_ids))
    )
    WHERE c.id = ANY(p_category_ids);

    GET DIAGNOSTICS v_rows_affected = ROW_COUNT;
    RETURN v_rows_affected;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION inherit_parameters_to_children(
    p_parent_id BIGINT
) RETURNS INT AS $$
DECLARE
    v_inheritable_numeric BIGINT[];
    v_inheritable_enum BIGINT[];
    v_descendants BIGINT[];
    v_rows_affected BIGINT := 0;
BEGIN
    SELECT array_agg(DISTINCT nid)
    INTO v_inheritable_numeric
    FROM unnest(coalesce((SELECT numeric_parameter_ids FROM classes WHERE id = p_parent_id), '{}')) AS nid
    JOIN range_parameters rp ON rp.id = nid
    WHERE rp.inheritable IS TRUE;

    SELECT array_agg(DISTINCT nid)
    INTO v_inheritable_enum
    FROM unnest(coalesce((SELECT enum_parameter_ids FROM classes WHERE id = p_parent_id), '{}')) AS nid
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
    p_parameter_ids BIGINT[],
    p_parent_ids BIGINT[]
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
    p_parameter_ids BIGINT[],
    p_parent_ids BIGINT[]
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
    UPDATE classes
    SET numeric_parameter_ids = array_remove(numeric_parameter_ids, OLD.id)
    WHERE numeric_parameter_ids @> ARRAY[OLD.id];

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- +goose StatementBegin
CREATE TRIGGER trg_before_delete_numeric_parameter
BEFORE DELETE ON range_parameters
FOR EACH ROW
EXECUTE FUNCTION cleanup_numeric_parameter_from_categories();
-- +goose StatementEnd


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION cleanup_enum_parameter_from_categories()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE classes
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
DROP TRIGGER IF EXISTS trg_before_delete_numeric_parameter ON range_parameters;
DROP FUNCTION IF EXISTS cleanup_enum_parameter_from_categories();
DROP FUNCTION IF EXISTS cleanup_numeric_parameter_from_categories();
DROP FUNCTION IF EXISTS remove_enum_parameters_from_descendants(BIGINT[], BIGINT[]);
DROP FUNCTION IF EXISTS remove_numeric_parameters_from_descendants(BIGINT[], BIGINT[]);
DROP FUNCTION IF EXISTS inherit_parameters_to_children(BIGINT);
DROP FUNCTION IF EXISTS get_descendants(BIGINT[]);
DROP FUNCTION IF EXISTS remove_enum_parameters_from_categories(BIGINT[], BIGINT[]);
DROP FUNCTION IF EXISTS add_enum_parameters_to_categories(BIGINT[], BIGINT[]);
DROP FUNCTION IF EXISTS remove_numeric_parameters_from_categories(BIGINT[], BIGINT[]);
DROP FUNCTION IF EXISTS add_numeric_parameters_to_categories(BIGINT[], BIGINT[]);
DROP TRIGGER IF EXISTS trg_enum_values_reorder_after_delete ON parameter_values;
DROP TRIGGER IF EXISTS trg_enum_values_reorder_before ON parameter_values;
DROP FUNCTION IF EXISTS fn_enum_values_reorder_after_delete;
DROP FUNCTION IF EXISTS fn_enum_values_reorder_before;