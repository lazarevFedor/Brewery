package entities

// Operation описывает технологическую операцию.
type Operation struct {
	ID           int64          `json:"id" info:"ID операции"`
	Name         string         `json:"name" info:"Название операции"`
	Description  string         `json:"description,omitempty" info:"Описание технологической операции"`
	ClassID      int64          `json:"operation_class_id" info:"ID класса технологической операции"`
	ProfessionID int64          `json:"profession_id" info:"ID требуемой профессии работника"`
	LaborCost    float64        `json:"labor_cost" info:"Трудозатратность в человеко-часах"`
	Parameters   map[string]any `json:"parameters,omitempty" info:"Технологические параметры операции"`
	StepNumber   int32          `json:"step_number,omitempty" info:"Номер операции в последовательности тех карты"`
	// Связанные сущности
	Equipment Equipment       `json:"equipment" info:"Данные оборудования"`
	Materials MaterialProduct `json:"materials" info:"Данные материала операции"`
}

// ProductTechCard описывает технологическую карту изделия.
type ProductTechCard struct {
	ProductID     int64             `json:"product_id"`
	TotalCost     float64           `json:"total_cost"`
	Operations    []Operation       `json:"operations"`
	MaterialNorms []MaterialProduct `json:"material_norms"`
}
