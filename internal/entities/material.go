package entities

type Material struct {
	MaterialID     int64
	ProductID      int64
	MaterialAmount float32
}

type MaterialProduct struct {
	Material
	Product
}

// MaterialTreeNode - узел дерева состава изделия (результат FIND_BOM).
type MaterialTreeNode struct {
	MaterialID     int64              `json:"material_id"`
	Name           string             `json:"name"`
	MaterialAmount float32            `json:"material_amount"`
	Units          string             `json:"units,omitempty"`
	Materials      []MaterialTreeNode `json:"materials"`
}

// MaterialNorm - суммарная норма расхода материала для изготовления изделия.
type MaterialNorm struct {
	MaterialID  int64   `json:"material_id"`
	Name        string  `json:"name"`
	TotalAmount float64 `json:"total_amount"`
	Units       string  `json:"units,omitempty"`
}
