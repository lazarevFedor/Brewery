package entities

type NumericParameter struct {
	ID         uint
	MinValue   int
	MaxValue   int
	FieldName  string
	EntityName string
}

type NumericParameters []NumericParameter

type EnumParameter struct {
	ID          uint
	EnumClassID uint
}

type EnumParameters []EnumParameter
