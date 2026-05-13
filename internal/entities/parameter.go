// Package entities содержит слой сущностей
package entities

// NumericParameter представляет собой структуру, описывающую численный параметр.
type NumericParameter struct {
	ID          uint   `json:"id,omitempty" info:"ID численного параметра"`
	MinValue    int    `json:"min_val,omitempty" info:"Минимальное значение параметра"`
	MaxValue    int    `json:"max_val,omitempty" info:"Максимальное значение параметра"`
	FieldName   string `json:"field_name,omitempty" info:"Имя поля сущности"`
	EntityName  string `json:"entity_name,omitempty" info:"Имя сущности, к которой относится параметр"`
	Inheritable bool   `json:"inheritable,omitempty" info:"Флаг, указывающий, может ли параметр наследоваться от родительской категории"`
}

// NumericParameters представляет собой срез численных параметров.
//
//easyjson:json
type NumericParameters []NumericParameter

// EnumParameter представляет собой структуру, описывающую параметр-перечисление.
type EnumParameter struct {
	ID          uint `json:"id,omitempty" info:"ID параметра-перечисления"`
	EnumClassID uint `json:"enum_class_id,omitempty" info:"ID класса перечисления"`
	Inheritable bool `json:"inheritable,omitempty" info:"Флаг, указывающий, может ли параметр наследоваться от родительской категории"`
}

// EnumParameters представляет собой срез параметров-перечислений.
//
//easyjson:json
type EnumParameters []EnumParameter

<<<<<<< feat/parameters_db
type Operation string

const (
	OpEqual        Operation = "="
	OpGreater      Operation = ">"
	OpGreaterEqual Operation = ">="
	OpLess         Operation = "<"
	OpLessEqual    Operation = "<="
	OpNotEqual     Operation = "!="
)

type FilterParameter struct {
	FieldName string    `json:"field_name,omitempty" info:"Имя поля сущности"`
	Operation Operation `json:"operation,omitempty" info:"Операция сравнения (eq, gt, ge, lt, le, ne)"`
	Value     float32   `json:"value,omitempty" info:"Значение для сравнения"`
}
=======
const (
	MissingType = iota
	NumericParameterType
	EnumParameterType
)
>>>>>>> dev
