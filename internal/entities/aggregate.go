// Package entities содержит слой сущностей
package entities

type Aggregate struct {
	ID                uint
	Name              string
	Description       string
	NumericParameters []int
	EnumParameters    []int
}

type Aggregates []Aggregate
