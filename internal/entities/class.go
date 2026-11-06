// Package entities содержит слой сущностей
package entities

// Class представляет собой структуру, описывающую класс изделия.
type Class struct {
	ID            int64   `json:"id,omitempty" info:"ID класса товара"`
	Name          string  `json:"name" info:"Название класса товара"`
	ParentID      int64   `json:"parent_id,omitempty" info:"ID класса меньшего уровня"`
	RangeParamIDs []int64 `json:"range_param_ids,omitempty" info:"ID числовых параметров классов"`
	EnumParamIDs  []int64 `json:"enum_param_ids,omitempty" info:"ID перечислимых параметров классов"`
}
