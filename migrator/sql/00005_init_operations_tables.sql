-- +goose Up

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS price NUMERIC(12, 2),
    ADD COLUMN IF NOT EXISTS is_purchased BOOLEAN NOT NULL DEFAULT FALSE,
    ADD CONSTRAINT chk_product_price_positive CHECK (price >= 0);

CREATE TABLE IF NOT EXISTS work_centers (
    equipment_name TEXT NOT NULL,
    work_center_name TEXT NOT NULL,
    PRIMARY KEY (equipment_name, work_center_name)
);

CREATE TABLE IF NOT EXISTS operations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    description TEXT,
    name TEXT NOT NULL,
    equipment_name TEXT NOT NULL,
    work_center_name TEXT NOT NULL,
    step_number INT NOT NULL,
    labor_cost NUMERIC(10, 3) NOT NULL DEFAULT 0,
    profession_id BIGINT NOT NULL REFERENCES classes (id) ON DELETE RESTRICT,
    input_product_id BIGINT NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    output_product_id BIGINT NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    CONSTRAINT fk_operations_work_center
        FOREIGN KEY (equipment_name, work_center_name)
        REFERENCES work_centers (equipment_name, work_center_name)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_operations_material
        FOREIGN KEY (input_product_id, output_product_id)
        REFERENCES materials (material_id, product_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_operation_name_not_empty CHECK (btrim(name) <> ''),
    CONSTRAINT chk_operation_step_number_positive CHECK (step_number > 0),
    CONSTRAINT chk_operation_labor_cost_positive CHECK (labor_cost >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_operations_output_product_step
ON operations USING btree (output_product_id, step_number);

CREATE INDEX IF NOT EXISTS idx_operations_work_center
ON operations USING btree (equipment_name, work_center_name);

CREATE INDEX IF NOT EXISTS idx_operations_profession_id
ON operations USING btree (profession_id);

CREATE INDEX IF NOT EXISTS idx_operations_input_product_id
ON operations USING btree (input_product_id);


-- +goose Down
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS work_centers;
ALTER TABLE products
    DROP CONSTRAINT IF EXISTS chk_product_price_positive,
    DROP COLUMN IF EXISTS is_purchased,
    DROP COLUMN IF EXISTS price;
