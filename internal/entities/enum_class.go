package entities

type EnumClass struct {
	ID         int    `json:"id,omitempty" info:"ID класса перечисления"`
	Type       string `json:"type,omitempty" info:"Тип класса перечисления"`
	EntityName string `json:"entity_name,omitempty" info:"Имя сущности, к которой относится класс перечисления"`
	FieldName  string `json:"field_name,omitempty" info:"Имя поля у сущности, к которому относится класс перечисления"`
}
