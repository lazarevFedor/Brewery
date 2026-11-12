package entities

// Operation описывает технологическую операцию преобразования входного продукта в выходной.
type Operation struct {
	ID              int64          `json:"id" info:"ID операции"`
	Name            string         `json:"name" info:"Название операции"`
	Description     string         `json:"description" info:"Описание операции"`
	ProfessionID    int64          `json:"profession_id" info:"ID требуемой профессии работника"`
	EquipmentID     int64          `json:"equipment_id" info:"ID оборудования для выполнения"`
	InputProductID  int64          `json:"input_product_id" info:"ID входного продукта"`
	OutputProductID int64          `json:"output_product_id" info:"ID выходного продукта"`
	LaborCost       float64        `json:"labor_cost" info:"Трудозатратность в человеко-часах"`
	Parameters      map[string]any `json:"parameters,omitempty" info:"Технологические параметры операции"`

	// Связанные сущности
	Equipment     *Equipment `json:"equipment,omitempty" info:"Данные оборудования"`
	InputProduct  *Product   `json:"input_product,omitempty" info:"Данные входного продукта"`
	OutputProduct *Product   `json:"output_product,omitempty" info:"Данные выходного продукта"`
}

// ProductTechCard описывает технологическую карту изделия.
type ProductTechCard struct {
	ProductID      int64             `json:"product_id"`
	Product        *Product          `json:"product,omitempty"`
	TotalLaborCost float64           `json:"total_labor_cost"`
	Operations     []Operation       `json:"operations"`
	MaterialNorms  []MaterialProduct `json:"material_norms"`
}
