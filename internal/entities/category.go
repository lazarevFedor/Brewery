// Package entities содержит слой сущностей
package entities

// ProductCategory представляет собой структуру, описывающую категорию изделия.
type ProductCategory struct {
	ID       int    `json:"id,omitempty" info:"ID категории товара"`
	Name     string `json:"name" info:"Название категории товара"`
	ParentID int    `json:"parent_id,omitempty" info:"ID категории меньшего уровня"`
}

// Products представляет собой срез категорий изделий.
//
//easyjson:json
type Products []ProductCategory
