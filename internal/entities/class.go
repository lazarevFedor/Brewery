// Package entities содержит слой сущностей
package entities

// Class представляет собой структуру, описывающую класс изделия.
type Class struct {
	ID           int    `json:"id,omitempty" info:"ID категории товара"`
	Name         string `json:"name" info:"Название категории товара"`
	ParentID     int    `json:"parent_id,omitempty" info:"ID категории меньшего уровня"`
	NumParamIDs  []int  `json:"num_param_ids,omitempty" info:"ID числовых параметров категории"`
	EnumParamIDs []int  `json:"enum_param_ids,omitempty" info:"ID перечислимых параметров категории"`
}

// Classes представляет собой срез классов изделий.
//
//easyjson:json
type Classes []Class
