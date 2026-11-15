// Package entities содержит слой сущностей
package entities

// Product представляет собой структуру, описывающую изделие.
type Product struct {
	ID             int64          `json:"id" info:"ID записи в базе данных"`
	Name           string         `json:"name" info:"Название изделия"`
	ProdClass      Class          `json:"class" info:"Информация о классе изделия"`
	BaseID         int64          `json:"base_id,omitempty" info:"ID базового изделия"`
	Parameters     map[string]any `json:"parameters,omitempty" info:"Список параметров изделия"`
	Unit           string         `json:"unit,omitempty" info:"Единица измерений"`
	Amount         float32        `json:"amount,omitempty" info:"Количество изделия"`
	IsFinalProduct bool           `json:"is_final_product,omitempty" info:"Флаг конечного готового продукта"`
}
