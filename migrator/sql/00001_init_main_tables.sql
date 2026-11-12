-- +goose Up
-- ==========================================
--                TABLES
-- ==========================================

CREATE TABLE IF NOT EXISTS classes(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    parent_id BIGINT REFERENCES classes (id) ON DELETE RESTRICT,
    numeric_parameter_ids BIGINT[],
    enum_parameter_ids BIGINT[]
);

CREATE TABLE IF NOT EXISTS products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    amount NUMERIC(10,3),
    units TEXT,
    class_id BIGINT NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
    base_id BIGINT REFERENCES products (id) ON DELETE SET NULL,
    CONSTRAINT chk_product_isnot_its_base CHECK (id != base_id),
    CONSTRAINT check_product_amount_positive CHECK (amount >= 0)
);

CREATE TABLE IF NOT EXISTS materials (
    material_id BIGINT REFERENCES products(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id) ON DELETE CASCADE,      
    material_amount NUMERIC(10,3) NOT NULL,
    PRIMARY KEY(material_id, product_id),
    CONSTRAINT chk_material_isnot_product CHECK (material_id != product_id)
);

CREATE TABLE IF NOT EXISTS change_history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    changes TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
--               CONSTRAINTS
-- ==========================================

-- Ограничение единственности корня
CREATE UNIQUE INDEX idx_classes_single_root 
ON classes ((true)) 
WHERE parent_id IS NULL;

-- ==========================================
--                INDEXES
-- ==========================================

-- Индексы для поиска по material_id и product_id
CREATE INDEX idx_materials_material_id ON materials(material_id);
CREATE INDEX idx_materials_product_id ON materials(product_id);

-- Индекс для поиска по parent_id
CREATE INDEX idx_classes_parent_id ON classes (parent_id);

-- Индекс для поиска товаров по категориям
CREATE INDEX idx_products_class_id ON products(class_id);

-- Индекс для поиска истории изменения товара
CREATE INDEX idx_change_history_product_id 
ON change_history(product_id);

CREATE INDEX idx_change_history_created_at
ON change_history
USING BRIN (created_at);




-- +goose Down
DROP TABLE IF EXISTS change_history;
DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS classes;
