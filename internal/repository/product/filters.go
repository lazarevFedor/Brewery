package product

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"fmt"
	"math"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

func productFilter(filter entities.FilterParameter) (sq.Sqlizer, error) {
	raw := []string{fmt.Sprint(filter.Value)}
	if filter.Operation == entities.OpIn {
		raw = strings.Split(raw[0], ",")
		for i := range raw {
			raw[i] = strings.TrimSpace(raw[i])
		}
	}
	textValues := make([]any, 0, len(raw))
	numbers := make([]any, 0, len(raw))
	for _, value := range raw {
		textValues = append(textValues, value)
		if number, err := strconv.ParseFloat(value, 64); err == nil && !math.IsNaN(number) && !math.IsInf(number, 0) {
			numbers = append(numbers, number)
		}
	}
	switch filter.Operation {
	case entities.OpEqual, entities.OpNotEqual, entities.OpIn:
	case entities.OpGreater, entities.OpGreaterEqual, entities.OpLess, entities.OpLessEqual:
		if len(numbers) == 0 {
			return nil, apperrors.BadRequest("ordered filters require a finite numeric value", nil)
		}
	default:
		return nil, apperrors.BadRequest("unsupported filter operation", nil)
	}

	rangeQuery := sq.Select("1").From("products_parameters pp").
		Join("range_parameters rp ON rp.id = pp.range_parameter").
		Where("pp.product_id = p.id").Where(sq.Eq{"rp.field_name": filter.FieldName}).
		Where(compareParameter("pp.numeric_value", filter.Operation, numbers))
	enumPredicates := sq.Or{
		sq.And{sq.Eq{"ep.enum_type": []string{"int", "float"}}, compareParameter("CASE WHEN ep.enum_type IN ('int', 'float') THEN pv.value_raw::numeric END", filter.Operation, numbers)},
	}
	if filter.Operation == entities.OpEqual || filter.Operation == entities.OpNotEqual || filter.Operation == entities.OpIn {
		enumPredicates = append(enumPredicates, sq.And{sq.Eq{"ep.enum_type": "string"}, compareParameter("pv.value_raw", filter.Operation, textValues)})
	}
	enumQuery := sq.Select("1").From("products_parameters pp").
		Join("enum_parameters ep ON ep.id = pp.enum_parameter_id").
		Join("parameter_values pv ON pv.id = pp.enum_value_id AND pv.enum_parameter_id = pp.enum_parameter_id").
		Where("pp.product_id = p.id").Where(sq.Eq{"ep.field_name": filter.FieldName}).Where(enumPredicates)
	matches := sq.Or{sq.Expr("EXISTS (?)", rangeQuery), sq.Expr("EXISTS (?)", enumQuery)}

	// Standard product fields remain searchable when the class has no parameter with that name.
	var builtin sq.Sqlizer
	switch filter.FieldName {
	case "id", "class_id", "base_id", "amount":
		builtin = compareParameter("p."+filter.FieldName, filter.Operation, numbers)
	case "name", "units":
		if filter.Operation == entities.OpEqual || filter.Operation == entities.OpNotEqual || filter.Operation == entities.OpIn {
			builtin = compareParameter("p."+filter.FieldName, filter.Operation, textValues)
		}
	}
	if builtin != nil {
		classRange := sq.Select("1").From("range_parameters rp").Where("rp.id = ANY(c.numeric_parameter_ids)").Where(sq.Eq{"rp.field_name": filter.FieldName})
		classEnum := sq.Select("1").From("enum_parameters ep").Where("ep.id = ANY(c.enum_parameter_ids)").Where(sq.Eq{"ep.field_name": filter.FieldName})
		matches = append(matches, sq.And{sq.Expr("NOT EXISTS (?)", classRange), sq.Expr("NOT EXISTS (?)", classEnum), builtin})
	}
	return matches, nil
}

func compareParameter(column string, operation entities.Operation, values []any) sq.Sqlizer {
	if len(values) == 0 {
		return sq.Expr("FALSE")
	}
	switch operation {
	case entities.OpEqual:
		return sq.Eq{column: values[0]}
	case entities.OpNotEqual:
		return sq.NotEq{column: values[0]}
	case entities.OpGreater:
		return sq.Gt{column: values[0]}
	case entities.OpGreaterEqual:
		return sq.GtOrEq{column: values[0]}
	case entities.OpLess:
		return sq.Lt{column: values[0]}
	case entities.OpLessEqual:
		return sq.LtOrEq{column: values[0]}
	case entities.OpIn:
		return sq.Eq{column: values}
	default:
		return sq.Expr("FALSE")
	}
}
