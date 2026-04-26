-- +goose Up

CREATE TABLE IF NOT EXISTS enum_classes (
    id SERIAL PRIMARY KEY,
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
    id SERIAL PRIMARY KEY,
    enum_class_id INT NOT NULL,

    value_raw VARCHAR(255) NOT NULL,   -- само значение

    position INT NOT NULL,
    CONSTRAINT chk_enum_values_position_positive CHECK (position > 0),
    CONSTRAINT uq_enum_values_class_position UNIQUE (enum_class_id, position) DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (enum_class_id) REFERENCES enum_classes(id) ON DELETE CASCADE
);

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
        PERFORM 1 FROM enum_values WHERE enum_class_id = NEW.enum_class_id FOR UPDATE;

        SELECT COUNT(*) INTO max_pos
        FROM enum_values
        WHERE enum_class_id = NEW.enum_class_id;

        IF NEW.position IS NULL OR NEW.position < 1 THEN
            NEW.position := max_pos + 1;
        ELSIF NEW.position > max_pos + 1 THEN
            NEW.position := max_pos + 1;
        END IF;

        UPDATE enum_values
        SET position = position + 1
        WHERE enum_class_id = NEW.enum_class_id
          AND position >= NEW.position;

        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        IF NEW.position IS NULL THEN
            NEW.position := OLD.position;
        END IF;

        IF NEW.enum_class_id = OLD.enum_class_id THEN
            PERFORM 1 FROM enum_values WHERE enum_class_id = NEW.enum_class_id FOR UPDATE;

            SELECT COUNT(*) INTO max_pos
            FROM enum_values
            WHERE enum_class_id = NEW.enum_class_id
              AND id <> OLD.id;

            IF NEW.position < 1 THEN
                NEW.position := 1;
            ELSIF NEW.position > max_pos + 1 THEN
                NEW.position := max_pos + 1;
            END IF;

            IF NEW.position < OLD.position THEN
                UPDATE enum_values
                SET position = position + 1
                WHERE enum_class_id = OLD.enum_class_id
                  AND id <> OLD.id
                  AND position >= NEW.position
                  AND position < OLD.position;
            ELSIF NEW.position > OLD.position THEN
                UPDATE enum_values
                SET position = position - 1
                WHERE enum_class_id = OLD.enum_class_id
                  AND id <> OLD.id
                  AND position > OLD.position
                  AND position <= NEW.position;
            END IF;

            RETURN NEW;
        END IF;

        -- Move between classes: compact old class and insert in new class position.
        PERFORM 1 FROM enum_values WHERE enum_class_id IN (OLD.enum_class_id, NEW.enum_class_id) FOR UPDATE;

        UPDATE enum_values
        SET position = position - 1
        WHERE enum_class_id = OLD.enum_class_id
          AND position > OLD.position;

        SELECT COUNT(*) INTO max_pos
        FROM enum_values
        WHERE enum_class_id = NEW.enum_class_id;

        IF NEW.position < 1 THEN
            NEW.position := 1;
        ELSIF NEW.position > max_pos + 1 THEN
            NEW.position := max_pos + 1;
        END IF;

        UPDATE enum_values
        SET position = position + 1
        WHERE enum_class_id = NEW.enum_class_id
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

    UPDATE enum_values
    SET position = position - 1
    WHERE enum_class_id = OLD.enum_class_id
      AND position > OLD.position;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_enum_values_reorder_before
BEFORE INSERT OR UPDATE ON enum_values
FOR EACH ROW
EXECUTE FUNCTION fn_enum_values_reorder_before();

CREATE TRIGGER trg_enum_values_reorder_after_delete
AFTER DELETE ON enum_values
FOR EACH ROW
EXECUTE FUNCTION fn_enum_values_reorder_after_delete();

-- +goose Down
DROP TRIGGER IF EXISTS trg_enum_values_reorder_after_delete ON enum_values;
DROP TRIGGER IF EXISTS trg_enum_values_reorder_before ON enum_values;
DROP FUNCTION IF EXISTS fn_enum_values_reorder_after_delete;
DROP FUNCTION IF EXISTS fn_enum_values_reorder_before;
DROP TABLE IF EXISTS enum_values;
DROP TABLE IF EXISTS enum_classes;
