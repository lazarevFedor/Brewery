package handlers

import "Brewery/internal/entities"

// enumParameter соответствует схеме EnumParameter в Swagger.
type enumParameter struct {
	ID          int    `json:"id"`
	EnumType    string `json:"enum_type"`
	FieldName   string `json:"field_name"`
	Unit        string `json:"unit,omitempty"`
	Inheritable bool   `json:"inheritable"`
}

func (p enumParameter) toEntity() entities.EnumClass {
	return entities.EnumClass{
		ID:          p.ID,
		Type:        p.EnumType,
		FieldName:   p.FieldName,
		Unit:        p.Unit,
		Inheritable: p.Inheritable,
	}
}

func enumParameterFromEntity(p entities.EnumClass) enumParameter {
	return enumParameter{
		ID:          p.ID,
		EnumType:    p.Type,
		FieldName:   p.FieldName,
		Unit:        p.Unit,
		Inheritable: p.Inheritable,
	}
}

// enumValue соответствует схемам EnumValue и EnumValueCreate в Swagger.
type enumValue struct {
	ID          int               `json:"id"`
	ParameterID int               `json:"parameter_id"`
	Value       any               `json:"value"`
	EnumType    entities.EnumType `json:"enum_type"`
	Position    int               `json:"position"`
}

func (v enumValue) toEntity() entities.EnumValue {
	return entities.EnumValue{
		ID:          v.ID,
		EnumClassID: v.ParameterID,
		Value:       v.Value,
		ValueType:   v.EnumType,
		Position:    v.Position,
	}
}

func enumValueFromEntity(v entities.EnumValue) enumValue {
	return enumValue{
		ID:          v.ID,
		ParameterID: v.EnumClassID,
		Value:       v.Value,
		EnumType:    v.ValueType,
		Position:    v.Position,
	}
}
