package entities

type ProductCategory struct {
	ID         int    `json:"category,omitempty" info:"ID категории товара"`
	Name       string `json:"name" info:"Название категории товара"`
	Level      int    `json:"level" info:"Уровень вложенности категории товара"`
	ParentID   int    `json:"parent_id,omitempty" info:"ID категории меньшего уровня"`
	ChildrenID []int  `json:"children_id,omitempty" info:"Список ID категории большего уровня"`
}
