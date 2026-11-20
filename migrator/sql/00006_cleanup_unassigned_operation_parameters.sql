-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cleanup_unassigned_operation_parameters()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF pg_trigger_depth() > 1 THEN
        RETURN NEW;
    END IF;

    DELETE FROM operations_parameters op
    USING operations o
    WHERE op.operation_id = o.id AND o.operation_class_id = NEW.id
      AND (
          (op.range_parameter IS NOT NULL
              AND NOT COALESCE(op.range_parameter = ANY(NEW.numeric_parameter_ids), FALSE))
          OR (op.enum_parameter_id IS NOT NULL
              AND NOT COALESCE(op.enum_parameter_id = ANY(NEW.enum_parameter_ids), FALSE))
      );
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_cleanup_unassigned_operation_parameters
AFTER UPDATE OF numeric_parameter_ids, enum_parameter_ids ON classes
FOR EACH ROW
WHEN (OLD.numeric_parameter_ids IS DISTINCT FROM NEW.numeric_parameter_ids
   OR OLD.enum_parameter_ids IS DISTINCT FROM NEW.enum_parameter_ids)
EXECUTE FUNCTION cleanup_unassigned_operation_parameters();

-- +goose Down
DROP TRIGGER trg_cleanup_unassigned_operation_parameters ON classes;
DROP FUNCTION cleanup_unassigned_operation_parameters();
