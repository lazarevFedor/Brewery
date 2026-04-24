package entities

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	EnumValueTypeString  EnumType = "string"
	EnumValueTypeInt     EnumType = "int"
	EnumValueTypeFloat   EnumType = "float"
	EnumValueTypePicture EnumType = "picture"
)

type EnumType string

//easyjson:skip
type EnumValueRow struct {
	ID          int
	EnumClassID int
	ValueRaw    string
	ValueType   EnumType
	Position    int
}

type EnumValue struct {
	ID          int      `json:"id,omitempty" info:"ID значения перечисления"`
	EnumClassID int      `json:"class_id,omitempty" info:"ID класса перечисления"`
	Value       any      `json:"value,omitempty" info:"Значение перечисления"`
	ValueType   EnumType `json:"enum_type,omitempty" info:"Тип класса перечисления"`
	Position    int      `json:"position,omitempty" info:"Позиция значения в перечислении"`
}

//easyjson:json
type EnumValues []EnumValue

func (e *EnumValue) ToRow() (*EnumValueRow, error) {
	if e == nil {
		return nil, errors.New("enum value is nil")
	}

	valueStr, err := enumValueToRaw(e.ValueType, e.Value)
	if err != nil {
		return nil, fmt.Errorf("convert enum value to raw: %w", err)
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
		return nil, errors.New("enum value row is nil")
	}

	value, err := enumValueFromRaw(e.ValueType, e.ValueRaw)
	if err != nil {
		return nil, fmt.Errorf("convert enum value from raw: %w", err)
	}

	return &EnumValue{
		ID:          e.ID,
		EnumClassID: e.EnumClassID,
		Value:       value,
		ValueType:   e.ValueType,
		Position:    e.Position,
	}, nil
}

func enumValueToRaw(valueType EnumType, value any) (string, error) {
	switch valueType {
	case EnumValueTypeString, EnumValueTypePicture:
		str, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("invalid enum value type %s: expected string, got %T", valueType, value)
		}

		return str, nil
	case EnumValueTypeInt:
		switch v := value.(type) {
		case int:
			return strconv.Itoa(v), nil
		default:
			return "", fmt.Errorf("invalid enum value type %s: expected integer, got %T", valueType, value)
		}
	case EnumValueTypeFloat:
		switch v := value.(type) {
		case float32:
			return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		default:
			return "", fmt.Errorf("invalid enum value type %s: expected float, got %T", valueType, value)
		}
	default:
		return "", fmt.Errorf("invalid enum value type: %s", valueType)
	}
}

func enumValueFromRaw(valueType EnumType, valueRaw string) (any, error) {
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
		value, err := strconv.ParseFloat(valueRaw, 32)
		if err != nil {
			return nil, fmt.Errorf("parse float value %s: %w", valueRaw, err)
		}

		return value, nil
	default:
		return nil, fmt.Errorf("invalid enum value type: %s", valueType)
	}
}
