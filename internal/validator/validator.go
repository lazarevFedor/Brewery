// Package validator содержит функции для валидации сущностей на соответствие заданным параметрам категорий.
//
//nolint:cyclop
package validator

import (
	"fmt"
	"reflect"
	"strings"

	"Brewery/internal/entities"
)

// ValidateBeerWithParams валидирует пиво на соответствие заданным параметрам у категории.
//
//nolint:funlen
func ValidateBeerWithParams(beer entities.Beer, numericParams []entities.NumericParameter, enumParams []entities.EnumParameter, getEnumValues func(classID uint) ([]entities.EnumValue, error), getEnumClass func(classID uint) (*entities.EnumClass, error)) error {
	val := reflect.ValueOf(beer)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	getField := func(fieldName string) (reflect.Value, bool) {
		t := val.Type()
		for i := range t.NumField() {
			f := t.Field(i)

			tag := strings.Split(f.Tag.Get("json"), ",")[0]
			if tag != "" && tag == fieldName {
				return val.Field(i), true
			}

			if strings.EqualFold(f.Name, fieldName) || strings.EqualFold(f.Name, strings.ToUpper(fieldName)) {
				return val.Field(i), true
			}
		}
		return reflect.Value{}, false
	}

	normalizeValue := func(rv reflect.Value) reflect.Value {
		for rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return reflect.Value{}
			}
			rv = rv.Elem()
		}
		return rv
	}

	for _, p := range numericParams {
		fv, ok := getField(p.FieldName)
		if !ok {
			continue
		}
		fv = normalizeValue(fv)
		if !fv.IsValid() {
			continue
		}

		var valFloat float64
		kind := fv.Kind()
		switch kind {
		case reflect.Float32, reflect.Float64:
			valFloat = fv.Float()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valFloat = float64(fv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			valFloat = float64(fv.Uint())
		case reflect.Invalid, reflect.Bool, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.String, reflect.Struct, reflect.UnsafePointer:
			continue
		default:
			continue
		}

		if p.MinValue != 0 {
			if valFloat < float64(p.MinValue) {
				return fmt.Errorf("field %s value %v less than min %d", p.FieldName, valFloat, p.MinValue)
			}
		}
		if p.MaxValue != 0 {
			if valFloat > float64(p.MaxValue) {
				return fmt.Errorf("field %s value %v greater than max %d", p.FieldName, valFloat, p.MaxValue)
			}
		}
	}

	for _, ep := range enumParams {
		cls, err := getEnumClass(ep.EnumClassID)
		if err != nil || cls == nil {
			if err != nil {
				return fmt.Errorf("failed to get enum class: %w", err)
			}
			continue
		}
		fv, ok := getField(cls.FieldName)
		if !ok {
			continue
		}
		fv = normalizeValue(fv)
		if !fv.IsValid() {
			continue
		}

		vals, err := getEnumValues(ep.EnumClassID)
		if err != nil {
			return fmt.Errorf("failed to get enum values: %w", err)
		}
		allowed := make(map[string]struct{}, len(vals))
		for _, v := range vals {
			allowed[fmt.Sprint(v.Value)] = struct{}{}
		}

		if _, ok = allowed[fmt.Sprint(fv.Interface())]; !ok {
			return fmt.Errorf("field %s value %v is not allowed", cls.FieldName, fv.Interface())
		}
	}

	return nil
}
