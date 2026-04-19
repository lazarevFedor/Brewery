package entities

import (
	"fmt"
	"strconv"
)

const (
	EnumValueTypeString  = "string"
	EnumValueTypeInt     = "int"
	EnumValueTypeFloat   = "float"
	EnumValueTypePicture = "picture"
)

//easyjson:skip
type EnumValueRow struct {
	ID          int
	EnumClassID int
	ValueRaw    string
	ValueType   string
	Position    int
}

type EnumValue struct {
	ID          int    `json:"id,omitempty" info:"ID значения перечисления"`
	EnumClassID int    `json:"enum_class_id,omitempty" info:"ID класса перечисления"`
	Value       any    `json:"value,omitempty" info:"Значение перечисления"`
	ValueType   string `json:"value_type,omitempty" info:"Тип значения перечисления"`
	Position    int    `json:"position,omitempty" info:"Позиция значения в перечислении"`
}

//easyjson:json
type EnumValues []EnumValue

func (e *EnumValue) ToRow() (*EnumValueRow, error) {
	if e == nil {
		return nil, fmt.Errorf("enum value is nil")
	}

	valueStr, err := enumValueToRaw(e.ValueType, e.Value)
	if err != nil {
		return nil, err
	}

	return &EnumValueRow{
		ID:          e.ID,
		EnumClassID: e.EnumClassID,
		ValueRaw:    valueStr,
		ValueType:   e.ValueType,
		Position:    e.Position,
	}, nil
}

func (e *EnumValueRow) FromRow() (*EnumValue, error) {
	if e == nil {
		return nil, fmt.Errorf("enum value row is nil")
	}

	value, err := enumValueFromRaw(e.ValueType, e.ValueRaw)
	if err != nil {
		return nil, err
	}

	return &EnumValue{
		ID:          e.ID,
		EnumClassID: e.EnumClassID,
		Value:       value,
		ValueType:   e.ValueType,
		Position:    e.Position,
	}, nil
}

func enumValueToRaw(valueType string, value any) (string, error) {
	switch valueType {
	case EnumValueTypeString, EnumValueTypePicture:
		str, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("invalid enum value type %q: expected string, got %T", valueType, value)
		}

		return str, nil
	case EnumValueTypeInt:
		switch v := value.(type) {
		case int:
			return strconv.Itoa(v), nil
		default:
			return "", fmt.Errorf("invalid enum value type %q: expected integer, got %T", valueType, value)
		}
	case EnumValueTypeFloat:
		switch v := value.(type) {
		case float32:
			return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
		default:
			return "", fmt.Errorf("invalid enum value type %q: expected float, got %T", valueType, value)
		}
	default:
		return "", fmt.Errorf("invalid enum value type: %s", valueType)
	}
}

func enumValueFromRaw(valueType, valueRaw string) (any, error) {
	switch valueType {
	case EnumValueTypeString, EnumValueTypePicture:
		return valueRaw, nil
	case EnumValueTypeInt:
		value, err := strconv.Atoi(valueRaw)
		if err != nil {
			return nil, fmt.Errorf("parse int value %q: %w", valueRaw, err)
		}

		return value, nil
	case EnumValueTypeFloat:
		value, err := strconv.ParseFloat(valueRaw, 64)
		if err != nil {
			return nil, fmt.Errorf("parse float value %q: %w", valueRaw, err)
		}

		return value, nil
	default:
		return nil, fmt.Errorf("invalid enum value type: %s", valueType)
	}
}
