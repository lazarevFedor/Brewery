// Package entities содержит слой сущностей
package entities

// RangeParameter представляет собой структуру, описывающую численный параметр.
type RangeParameter struct {
	ID          int64  `json:"id" info:"ID численного параметра"`
	MinValue    int    `json:"min_val" info:"Минимальное значение параметра"`
	MaxValue    int    `json:"max_val" info:"Максимальное значение параметра"`
	FieldName   string `json:"field_name" info:"Имя поля сущности"`
	Inheritable bool   `json:"inheritable" info:"Флаг, указывающий, может ли параметр наследоваться от родительской категории"`
	Unit        string `json:"unit,omitempty" info:"Единица измерений"`
}

type Operation string

const (
	OpEqual        Operation = "="
	OpGreater      Operation = ">"
	OpGreaterEqual Operation = ">="
	OpLess         Operation = "<"
	OpLessEqual    Operation = "<="
	OpNotEqual     Operation = "!="
	OpIn           Operation = "IN"
)

type FilterParameter struct {
	FieldName string    `json:"field_name,omitempty" info:"Имя поля сущности"`
	Operation Operation `json:"operation,omitempty" info:"Операция сравнения (eq, gt, ge, lt, le, ne)"`
	Value     any       `json:"value,omitempty" info:"Значение для сравнения"`
}

const (
	MissingType = iota
	RangeParameterType
	EnumParameterType
)
