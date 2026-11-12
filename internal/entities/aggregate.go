// Package entities содержит слой сущностей
package entities

// Aggregate представляет собой структуру, описывающую агрегат.
type Aggregate struct {
	ID                int64   `json:"id,omitempty" info:"ID агрегата"`
	Name              string  `json:"name,omitempty" info:"Имя агрегата"`
	Description       string  `json:"description,omitempty" info:"Описание товара"`
	NumericParameters []int64 `json:"numeric_param_ids,omitempty" info:"Список численный параметров"`
	EnumParameters    []int64 `json:"enum_param_ids,omitempty" info:"Список параметров-перечислений"`
}
