// Package dto хранит в себе API контракты
package dto


type CategoryBase struct {
	Name     string `json:"name" info:"Название категории товара"`
}

type Category struct {
	CategoryBase
	
	ID       int    `json:"id,omitempty" info:"ID категории товара"`
	ParentID int    `json:"parent_id,omitempty" info:"ID категории меньшего уровня"`
}

type CategoryCreate struct {
	CategoryBase

	ParentID int    `json:"parent_id,omitempty" info:"ID категории меньшего уровня"`
}

type CategoryUpdate struct {
	CategoryBase

	ParentID int    `json:"parent_id,omitempty" info:"ID категории меньшего уровня"`
}