package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"
	"sort"
)

// resolveClassParameters преобразует API-значения в проверенные ссылки на параметры класса.
func resolveClassParameters(ctx context.Context, classRepo ClassParametersRepository, enumRepo EnumValuesRepository, classID int64, parameters map[string]any) ([]entities.ParameterValue, map[string]any, error) {
	resolved := make([]entities.ParameterValue, 0, len(parameters))
	canonical := make(map[string]any, len(parameters))
	if len(parameters) == 0 {
		return resolved, canonical, nil
	}
	ranges, enums, err := classRepo.GetParameters(ctx, classID, entities.MissingType)
	if err != nil {
		return nil, nil, fmt.Errorf("get class parameters: %w", err)
	}
	rangeByField := make(map[string]entities.RangeParameter)
	enumByField := make(map[string]entities.EnumClass)
	counts := make(map[string]int)
	for _, parameter := range ranges {
		rangeByField[parameter.FieldName] = parameter
		counts[parameter.FieldName]++
	}
	for _, parameter := range enums {
		enumByField[parameter.FieldName] = parameter
		counts[parameter.FieldName]++
	}
	fields := make([]string, 0, len(parameters))
	for field := range parameters {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		value := parameters[field]
		if counts[field] == 0 {
			return nil, nil, apperrors.BadRequest(fmt.Sprintf("parameter %q is not allowed for class %d", field, classID), nil)
		}
		if counts[field] > 1 {
			return nil, nil, apperrors.BadRequest(fmt.Sprintf("parameter field %q is ambiguous in class %d", field, classID), nil)
		}
		if raw, ok := value.(json.Number); ok {
			if integer, intErr := raw.Int64(); intErr == nil {
				value = integer
			} else {
				number, floatErr := raw.Float64()
				if floatErr != nil {
					return nil, nil, apperrors.BadRequest(fmt.Sprintf("parameter %q is out of range", field), floatErr)
				}
				value = number
			}
		}
		if parameter, ok := rangeByField[field]; ok {
			number, ok := toFloat64Value(value)
			if !ok || math.IsNaN(number) || math.IsInf(number, 0) {
				return nil, nil, apperrors.BadRequest(fmt.Sprintf("parameter %q must be a finite number", field), nil)
			}
			if number < float64(parameter.MinValue) || number > float64(parameter.MaxValue) {
				return nil, nil, apperrors.BadRequest(fmt.Sprintf("parameter %q must be between %d and %d", field, parameter.MinValue, parameter.MaxValue), nil)
			}
			if math.Abs(number) > 99999.999 || math.Abs(number*1000-math.Round(number*1000)) > 1e-6 {
				return nil, nil, apperrors.BadRequest(fmt.Sprintf("parameter %q must fit NUMERIC(8, 3)", field), nil)
			}
			id := parameter.ID
			resolved = append(resolved, entities.ParameterValue{FieldName: field, RangeParameterID: &id, NumericValue: &number})
			canonical[field] = number
			continue
		}
		parameter := enumByField[field]
		incoming := entities.EnumValue{Value: value, ValueType: entities.EnumType(parameter.Type)}
		incomingRow, err := incoming.ToRow()
		if err != nil {
			return nil, nil, apperrors.BadRequest(fmt.Sprintf("invalid value for parameter %q: %v", field, err), err)
		}
		allowed, err := enumRepo.GetEnumValuesByParameterID(ctx, uint(parameter.ID))
		if err != nil {
			return nil, nil, fmt.Errorf("get allowed values for %q: %w", field, err)
		}
		var matched *entities.EnumValue
		for i := range allowed {
			row, err := allowed[i].ToRow()
			if err != nil {
				return nil, nil, fmt.Errorf("decode allowed value for %q: %w", field, err)
			}
			if row.ValueRaw == incomingRow.ValueRaw {
				matched = &allowed[i]
				break
			}
		}
		if matched == nil {
			return nil, nil, apperrors.BadRequest(fmt.Sprintf("value %v is not allowed for parameter %q", value, field), nil)
		}
		parameterID, valueID := int64(parameter.ID), int64(matched.ID)
		resolved = append(resolved, entities.ParameterValue{FieldName: field, EnumParameterID: &parameterID, EnumValueID: &valueID})
		canonical[field] = matched.Value
	}
	return resolved, canonical, nil
}

func prepareParameterPatch(currentClassID, classID int64, current, patch map[string]any) (entities.ParameterChanges, map[string]any) {
	changes := entities.ParameterChanges{ExpectedClassID: currentClassID}
	candidate := make(map[string]any, len(patch))
	if classID != currentClassID {
		changes.Replace = true
		changes.ExpectedParameters = maps.Clone(current)
		maps.Copy(candidate, current)
	}
	for field, value := range patch {
		changes.Fields = append(changes.Fields, field)
		if value == nil {
			delete(candidate, field)
		} else {
			candidate[field] = value
		}
	}
	slices.Sort(changes.Fields)
	return changes, candidate
}
