// Package repository_test содержит тесты для слоя repository
package repository_test

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// testEnumClasses содержит набор тестовых классов перечислений для различных сценариев тестирования.
	testEnumClasses = []entities.EnumClass{
		{
			Type:       "int",
			EntityName: "beers",
			FieldName:  "amount",
			Unit:       "ml",
			IsActive:   true,
		},
		{
			Type:       "int",
			EntityName: "beers",
			FieldName:  "amount",
			Unit:       "L",
			IsActive:   false,
		},
		{
			Type:       "string",
			EntityName: "country",
			FieldName:  "name",
			IsActive:   true,
		},
		{
			Type:       "string",
			EntityName: "country",
			FieldName:  "name",
			IsActive:   true,
		},
		{
			Type:       "float",
			EntityName: "beers",
			FieldName:  "rating",
			IsActive:   true,
		},
	}

	// testEnumValues содержит набор тестовых значений перечислений для различных сценариев тестирования.
	testEnumValues = []entities.EnumValue{
		{
			Value:     10,
			ValueType: "int",
			Position:  1,
		},
		{
			Value:     20,
			ValueType: "int",
			Position:  2,
		},
		{
			Value:     30,
			ValueType: "int",
			Position:  3,
		},
		{
			Value:     5.0,
			ValueType: "float",
			Position:  1,
		},
		{
			Value:     "Germany",
			ValueType: "string",
			Position:  1,
		},
	}
)

// TestEnumRepository_InsertEnumClass содержит тесты для метода InsertEnumClass репозитория EnumPostgres.
func TestEnumRepository_InsertEnumClass(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка класса перечисления", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumClasses, err := enumRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.NoError(t, err)
		require.Len(t, enumClasses, 1)
		require.Equal(t, testEnumClasses[0].Type, enumClasses[0].Type)
		require.Equal(t, testEnumClasses[0].EntityName, enumClasses[0].EntityName)
		require.Equal(t, testEnumClasses[0].FieldName, enumClasses[0].FieldName)
		require.Equal(t, testEnumClasses[0].IsActive, enumClasses[0].IsActive)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Вставка двух классов перечислений одного типа", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumID, err = enumRepo.InsertEnumClass(ctx, testEnumClasses[1])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumClasses, err := enumRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.NoError(t, err)
		require.Len(t, enumClasses, 2)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Вставка двух активных классов перечислений одного поля", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[2])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumID, err = enumRepo.InsertEnumClass(ctx, testEnumClasses[3])
		require.Error(t, err)
		require.Zero(t, enumID)

		enumClasses, err := enumRepo.GetEnumClasses(ctx, testEnumClasses[2].EntityName, testEnumClasses[2].FieldName)
		require.NoError(t, err)
		require.Len(t, enumClasses, 1)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Вставка с использованием неинициализированного пула коннектов", func(t *testing.T) {
		uninitializedRepo := &repository.EnumPostgres{}

		enumID, err := uninitializedRepo.InsertEnumClass(ctx, testEnumClasses[2])
		require.Error(t, err)
		require.Zero(t, enumID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})
}

// TestEnumRepository_UpdateEnumClass содержит тесты для метода UpdateEnumClass репозитория EnumPostgres.
func TestEnumRepository_UpdateEnumClass(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное обновление класса перечисления", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		updates := map[string]any{
			"enum_type": "string",
			"unit":      nil,
			"is_active": false,
		}

		err = enumRepo.UpdateEnumClass(ctx, enumID, updates)
		require.NoError(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Обновление класса перечисления на активный при уже с существующем активном классе", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumID, err = enumRepo.InsertEnumClass(ctx, testEnumClasses[1])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		updates := map[string]any{
			"is_active": true,
		}

		err = enumRepo.UpdateEnumClass(ctx, enumID, updates)
		require.Error(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Обновление с использованием неинициализированного пула коннектов", func(t *testing.T) {
		uninitializedRepo := &repository.EnumPostgres{}

		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		updates := map[string]any{
			"enum_type": "string",
			"unit":      nil,
			"is_active": false,
		}

		err = uninitializedRepo.UpdateEnumClass(ctx, enumID, updates)
		require.Error(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})
}

// TestEnumRepository_DeleteEnumClassByID содержит тесты для метода DeleteEnumClassByID репозитория EnumPostgres.
func TestEnumRepository_DeleteEnumClassByID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное удаление класса перечисления", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumClasses, err := enumRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.NoError(t, err)
		require.Len(t, enumClasses, 1)

		err = enumRepo.DeleteEnumClassByID(ctx, enumID)
		require.NoError(t, err)

		enumClasses, err = enumRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.NoError(t, err)
		require.Empty(t, enumClasses)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Удаление несуществующего класса перечисления", func(t *testing.T) {
		enumClasses, err := enumRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.NoError(t, err)
		require.Empty(t, enumClasses)

		err = enumRepo.DeleteEnumClassByID(ctx, 999)
		require.NoError(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Удаление класса перечисления с использованием неинициализированного пула коннектов", func(t *testing.T) {
		uninitializedRepo := &repository.EnumPostgres{}

		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumClasses, err := enumRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.NoError(t, err)
		require.Len(t, enumClasses, 1)

		err = uninitializedRepo.DeleteEnumClassByID(ctx, enumID)
		require.Error(t, err)

		enumClasses, err = enumRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.NoError(t, err)
		require.Len(t, enumClasses, 1)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})
}

// TestEnumRepository_GetEnumClasses содержит тесты для метода GetEnumClasses репозитория EnumPostgres.
func TestEnumRepository_GetEnumClasses(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное получение классов перечисления", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumID, err = enumRepo.InsertEnumClass(ctx, testEnumClasses[1])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumClasses, err := enumRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.NoError(t, err)
		require.Len(t, enumClasses, 2)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Получение классов перечисления с использованием неинициализированного пула коннектов", func(t *testing.T) {
		uninitializedRepo := &repository.EnumPostgres{}

		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumID, err = enumRepo.InsertEnumClass(ctx, testEnumClasses[1])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		enumClasses, err := uninitializedRepo.GetEnumClasses(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName)
		require.Error(t, err)
		require.Empty(t, enumClasses)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})
}

// TestEnumRepository_InsertEnumValue содержит тесты для метода InsertEnumValue репозитория EnumPostgres.
//
//nolint:funlen
func TestEnumRepository_InsertEnumValue(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка значений перечисления", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		testVal := testEnumValues[0]
		testVal.EnumClassID = int(enumID)

		valueID, err := enumRepo.InsertEnumValue(ctx, testVal)
		require.NoError(t, err)
		require.NotZero(t, valueID)

		enumValues, err := enumRepo.GetEnumValues(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName, testVal.ValueType)
		require.NoError(t, err)
		require.Len(t, enumValues, 1)
		require.Equal(t, testVal.ValueType, enumValues[0].ValueType)
		require.Equal(t, testVal.Value, enumValues[0].Value)
		require.Equal(t, testVal.Position, enumValues[0].Position)

		testVal = testEnumValues[1]
		testVal.EnumClassID = int(enumID)

		valueID, err = enumRepo.InsertEnumValue(ctx, testVal)
		require.NoError(t, err)
		require.NotZero(t, valueID)

		enumValues, err = enumRepo.GetEnumValues(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName, testVal.ValueType)
		require.NoError(t, err)
		require.Len(t, enumValues, 2)
		require.Equal(t, testVal.ValueType, enumValues[1].ValueType)
		require.Equal(t, testVal.Value, enumValues[1].Value)
		require.Equal(t, testVal.Position, enumValues[1].Position)

		testVal = testEnumValues[2]
		testVal.EnumClassID = int(enumID)

		valueID, err = enumRepo.InsertEnumValue(ctx, testVal)
		require.NoError(t, err)
		require.NotZero(t, valueID)

		enumValues, err = enumRepo.GetEnumValues(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName, testVal.ValueType)
		require.NoError(t, err)
		require.Len(t, enumValues, 3)
		require.Equal(t, testVal.ValueType, enumValues[2].ValueType)
		require.Equal(t, testVal.Value, enumValues[2].Value)
		require.Equal(t, testVal.Position, enumValues[2].Position)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Вставка значения перечисления с использованием неинициализированного пула коннектов", func(t *testing.T) {
		uninitializedRepo := &repository.EnumPostgres{}

		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[0])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		testVal := testEnumValues[0]
		testVal.EnumClassID = int(enumID)

		valueID, err := uninitializedRepo.InsertEnumValue(ctx, testVal)
		require.Error(t, err)
		require.Zero(t, valueID)

		enumValues, err := enumRepo.GetEnumValues(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName, testVal.ValueType)
		require.NoError(t, err)
		require.Empty(t, enumValues)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Вставка значения перечисления с несуществующим классом", func(t *testing.T) {
		testVal := testEnumValues[0]
		testVal.EnumClassID = 999

		valueID, err := enumRepo.InsertEnumValue(ctx, testVal)
		require.Error(t, err)
		require.Zero(t, valueID)

		enumValues, err := enumRepo.GetEnumValues(ctx, testEnumClasses[0].EntityName, testEnumClasses[0].FieldName, testVal.ValueType)
		require.NoError(t, err)
		require.Empty(t, enumValues)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_values")
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Вставка значения перечисления типа float", func(t *testing.T) {
		enumID, err := enumRepo.InsertEnumClass(ctx, testEnumClasses[4])
		require.NoError(t, err)
		require.NotZero(t, enumID)

		testVal := testEnumValues[3]
		testVal.EnumClassID = int(enumID)

		valueID, err := enumRepo.InsertEnumValue(ctx, testVal)
		require.NoError(t, err)
		require.NotZero(t, valueID)

		enumValues, err := enumRepo.GetEnumValues(ctx, testEnumClasses[4].EntityName, testEnumClasses[4].FieldName, testVal.ValueType)
		require.NoError(t, err)
		require.Len(t, enumValues, 1)
		require.Equal(t, testVal.ValueType, enumValues[0].ValueType)
		require.Equal(t, testVal.Value, enumValues[0].Value)
		require.Equal(t, testVal.Position, enumValues[0].Position)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_values")
			cleanDB(t, ctx, "enum_classes")
		})
	})
}

// TestEnumRepository_GetEnumValues содержит тесты для метода GetEnumValues репозитория EnumPostgres.
func TestEnumRepository_GetEnumValues(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное получение значений перечисления", func(t *testing.T) {
		intClass := entities.EnumClass{
			Type:       "int",
			EntityName: "enum_get_beers",
			FieldName:  "amount",
			Unit:       "ml",
			IsActive:   true,
		}
		floatClass := entities.EnumClass{
			Type:       "float",
			EntityName: "enum_get_beers",
			FieldName:  "amount",
			Unit:       "L",
			IsActive:   false,
		}

		intClassID, err := enumRepo.InsertEnumClass(ctx, intClass)
		require.NoError(t, err)
		require.NotZero(t, intClassID)

		floatClassID, err := enumRepo.InsertEnumClass(ctx, floatClass)
		require.NoError(t, err)
		require.NotZero(t, floatClassID)

		firstVal := entities.EnumValue{EnumClassID: int(intClassID), Value: 10, ValueType: entities.EnumValueTypeInt, Position: 1}
		firstValID, err := enumRepo.InsertEnumValue(ctx, firstVal)
		require.NoError(t, err)
		require.NotZero(t, firstValID)

		secondVal := entities.EnumValue{EnumClassID: int(floatClassID), Value: 4.5, ValueType: entities.EnumValueTypeFloat, Position: 2}
		secondValID, err := enumRepo.InsertEnumValue(ctx, secondVal)
		require.NoError(t, err)
		require.NotZero(t, secondValID)

		intValues, err := enumRepo.GetEnumValues(ctx, intClass.EntityName, intClass.FieldName, entities.EnumType(intClass.Type))
		require.NoError(t, err)
		require.Len(t, intValues, 1)
		require.Equal(t, firstValID, uint(intValues[0].ID))
		require.Equal(t, firstVal.ValueType, intValues[0].ValueType)
		require.Equal(t, firstVal.Value, intValues[0].Value)
		require.Equal(t, firstVal.Position, intValues[0].Position)

		floatValues, err := enumRepo.GetEnumValues(ctx, floatClass.EntityName, floatClass.FieldName, entities.EnumType(floatClass.Type))
		require.NoError(t, err)
		require.Len(t, floatValues, 1)
		require.Equal(t, secondValID, uint(floatValues[0].ID))
		require.Equal(t, secondVal.ValueType, floatValues[0].ValueType)
		require.Equal(t, secondVal.Value, floatValues[0].Value)
		require.Equal(t, secondVal.Position, floatValues[0].Position)

		emptyValues, err := enumRepo.GetEnumValues(ctx, intClass.EntityName, intClass.FieldName, entities.EnumValueTypeString)
		require.NoError(t, err)
		require.Empty(t, emptyValues)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_values")
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Получение значений перечисления с использованием неинициализированного пула коннектов", func(t *testing.T) {
		uninitializedRepo := &repository.EnumPostgres{}

		values, err := uninitializedRepo.GetEnumValues(ctx, "beers", "amount", entities.EnumValueTypeInt)
		require.Error(t, err)
		require.Empty(t, values)
	})
}

// TestEnumRepository_UpdateEnumValue содержит тесты для метода UpdateEnumValue репозитория EnumPostgres.
func TestEnumRepository_UpdateEnumValue(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное обновление значения перечисления", func(t *testing.T) {
		enumClass := entities.EnumClass{Type: string(entities.EnumValueTypeInt), EntityName: "enum_update_beers_value", FieldName: "amount", Unit: "ml", IsActive: true}

		enumID, err := enumRepo.InsertEnumClass(ctx, enumClass)
		require.NoError(t, err)
		require.NotZero(t, enumID)

		testVal := entities.EnumValue{EnumClassID: int(enumID), Value: 10, ValueType: entities.EnumValueTypeInt, Position: 1}
		valueID, err := enumRepo.InsertEnumValue(ctx, testVal)
		require.NoError(t, err)
		require.NotZero(t, valueID)

		updates := map[string]any{
			"value_raw": 20,
			"position": 2,
		}
		err = enumRepo.UpdateEnumValue(ctx, valueID, updates)
		require.NoError(t, err)

		values, err := enumRepo.GetEnumValues(ctx, enumClass.EntityName, enumClass.FieldName, entities.EnumValueTypeInt)
		require.NoError(t, err)
		require.Len(t, values, 1)
		require.Equal(t, valueID, uint(values[0].ID))
		require.Equal(t, 20, values[0].Value)
		require.Equal(t, updates["position"], values[0].Position)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_values")
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Успешное обновление только позиции значения перечисления", func(t *testing.T) {
		enumClass := entities.EnumClass{Type: string(entities.EnumValueTypeInt), EntityName: "enum_update_beers_position", FieldName: "amount", Unit: "L", IsActive: true}

		enumID, err := enumRepo.InsertEnumClass(ctx, enumClass)
		require.NoError(t, err)
		require.NotZero(t, enumID)

		testVal := entities.EnumValue{EnumClassID: int(enumID), Value: 30, ValueType: entities.EnumValueTypeInt, Position: 1}
		valueID, err := enumRepo.InsertEnumValue(ctx, testVal)
		require.NoError(t, err)
		require.NotZero(t, valueID)

		updates := map[string]any{
			"value_raw": nil,
			"position": 5,
		}
		err = enumRepo.UpdateEnumValue(ctx, valueID, updates)
		require.NoError(t, err)

		values, err := enumRepo.GetEnumValues(ctx, enumClass.EntityName, enumClass.FieldName, entities.EnumValueTypeInt)
		require.NoError(t, err)
		require.Len(t, values, 1)
		require.Equal(t, valueID, uint(values[0].ID))
		require.Equal(t, testVal.Value, values[0].Value)
		require.Equal(t, updates["position"], values[0].Position)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_values")
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Обновление значения перечисления с использованием неинициализированного пула коннектов", func(t *testing.T) {
		uninitializedRepo := &repository.EnumPostgres{}

		enumClass := entities.EnumClass{Type: string(entities.EnumValueTypeInt), EntityName: "enum_update_uninit", FieldName: "amount", Unit: "kg", IsActive: true}

		enumID, err := enumRepo.InsertEnumClass(ctx, enumClass)
		require.NoError(t, err)
		require.NotZero(t, enumID)

		testVal := entities.EnumValue{EnumClassID: int(enumID), Value: 10, ValueType: entities.EnumValueTypeInt, Position: 1}
		valueID, err := enumRepo.InsertEnumValue(ctx, testVal)
		require.NoError(t, err)
		require.NotZero(t, valueID)

		updates := map[string]any{
			"value_raw": 20,
			"position": 2,
		}
		err = uninitializedRepo.UpdateEnumValue(ctx, valueID, updates)
		require.Error(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_values")
			cleanDB(t, ctx, "enum_classes")
		})
	})
}

// TestEnumRepository_DeleteEnumValueByID содержит тесты для метода DeleteEnumValueByID репозитория EnumPostgres.
func TestEnumRepository_DeleteEnumValueByID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное удаление значения перечисления", func(t *testing.T) {
		enumClass := entities.EnumClass{Type: string(entities.EnumValueTypeInt), EntityName: "enum_delete_beers", FieldName: "amount", Unit: "ml", IsActive: true}

		enumID, err := enumRepo.InsertEnumClass(ctx, enumClass)
		require.NoError(t, err)
		require.NotZero(t, enumID)

		testVal := entities.EnumValue{EnumClassID: int(enumID), Value: 10, ValueType: entities.EnumValueTypeInt, Position: 1}
		valueID, err := enumRepo.InsertEnumValue(ctx, testVal)
		require.NoError(t, err)
		require.NotZero(t, valueID)

		values, err := enumRepo.GetEnumValues(ctx, enumClass.EntityName, enumClass.FieldName, entities.EnumValueTypeInt)
		require.NoError(t, err)
		require.Len(t, values, 1)

		err = enumRepo.DeleteEnumValueByID(ctx, valueID)
		require.NoError(t, err)

		values, err = enumRepo.GetEnumValues(ctx, enumClass.EntityName, enumClass.FieldName, entities.EnumValueTypeInt)
		require.NoError(t, err)
		require.Empty(t, values)

		t.Cleanup(func() {
			cleanDB(t, ctx, "enum_values")
			cleanDB(t, ctx, "enum_classes")
		})
	})

	t.Run("Удаление несуществующего значения перечисления", func(t *testing.T) {
		err := enumRepo.DeleteEnumValueByID(ctx, 999)
		require.NoError(t, err)
	})

	t.Run("Удаление значения перечисления с использованием неинициализированного пула коннектов", func(t *testing.T) {
		uninitializedRepo := &repository.EnumPostgres{}

		err := uninitializedRepo.DeleteEnumValueByID(ctx, 999)
		require.Error(t, err)
	})
}
