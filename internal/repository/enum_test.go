// Package repository_test содержит тесты для слоя repository
package repository_test

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
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
	}
)

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
