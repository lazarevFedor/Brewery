// Package entities содержит слой сущностей
package entities

// Class представляет собой структуру, описывающую класс изделия.
type Class struct {
	ID           int64   `json:"id,omitempty" info:"ID категории товара"`
	Name         string  `json:"name" info:"Название категории товара"`
	ParentID     int64  `json:"parent_id,omitempty" info:"ID категории меньшего уровня"`
	NumParamIDs  []int64 `json:"num_param_ids,omitempty" info:"ID числовых параметров категории"`
	EnumParamIDs []int64 `json:"enum_param_ids,omitempty" info:"ID перечислимых параметров категории"`
}

