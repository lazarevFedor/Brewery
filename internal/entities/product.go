// Package entities содержит слой сущностей
package entities

// Product представляет собой структуру, описывающую изделие.
type Product struct {
	ID         uint           `json:"id" info:"ID записи в базе данных"`
	Name       string         `json:"name" info:"Название изделия"`
	Category   Class          `json:"class" info:"Информация о классе изделия"`
	BaseID     int            `json:"base_id,omitempty" info:"ID базового изделия"`
	Parameters map[string]any `json:"parameters,omitempty" info:"Список параметров изделия"`
}

// Products представляет собой срез изделий.
//
//easyjson:json
type Products []Product
