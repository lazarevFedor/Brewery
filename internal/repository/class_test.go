// Package repository_test содержит тесты для слоя repository
package repository_test

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCtg - глобальная переменная для хранения тестовой категории, которая будет использоваться в нескольких тестах.
var testCtg = entities.ProductCategory{Name: "test"}

// TestCategoryRepository_InsertCategory проверяет, что метод InsertCategory корректно вставляет категорию в базу данных и возвращает её ID.
func TestCategoryRepository_InsertCategory(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка", func(t *testing.T) {
		err := seedTestData(ctx)
		require.NoError(t, err)

		rootID, err := ctgRepo.GetCategoryID(ctx, nil, "test_category")
		require.NoError(t, err)
		require.NotZero(t, rootID)

		testCtg = entities.ProductCategory{Name: "test", ParentID: int(rootID)}

		ctgID, err := ctgRepo.InsertCategory(ctx, nil, testCtg)
		require.NoError(t, err)
		require.NotZero(t, ctgID)

		category, err := ctgRepo.GetCategoryByID(ctx, ctgID)
		require.NoError(t, err, "GetCategories Error")
		require.NotNil(t, category)
		require.Equal(t, testCtg.Name, category.Name)

		t.Cleanup(func() {
			cleanDB(t, ctx, "product_categories")
		})
	})

	t.Run("Дублирование категории", func(t *testing.T) {
		// Заполняем первый раз
		testCtg := entities.ProductCategory{Name: "test"}
		ctgID, err := ctgRepo.InsertCategory(ctx, nil, testCtg)

		require.NoError(t, err)
		assert.NotZero(t, ctgID)

		// Заполняем второй раз
		_, err = ctgRepo.InsertCategory(ctx, nil, testCtg)

		if err == nil {
			t.Error("Ожидалась ошибка уникальности, но запись создалась")
		}

		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if pgErr, ok := errors.AsType[*pgconn.PgError](appErr.Err); ok {
				if pgErr.Code != "23505" {
					t.Errorf("ожидался код ошибки 23505, получили %s", pgErr.Code)
				}
			} else {
				t.Errorf("ожидалась ошибка pgconn.PgError, получили %T", appErr)
			}
		} else {
			t.Errorf("Ожидалась ошибка типа AppError")
		}

		t.Cleanup(func() {
			cleanDB(t, ctx, "product_categories")
		})
	})
}

// TestCategoryRepository_GetCategories проверяет, что метод GetCategories корректно возвращает список всех категорий из базы данных, а также обрабатывает случай, когда база данных пуста.
func TestCategoryRepository_GetCategories(t *testing.T) {
	ctx := t.Context()

	t.Run("Пустая БД", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")

		ctgs, err := ctgRepo.GetCategories(ctx)
		assert.Empty(t, ctgs)
		require.NoError(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "product_categories")
		})
	})

	t.Run("Успешное нахождение 1 категории", func(t *testing.T) {
		testCtg := entities.ProductCategory{Name: "test"}
		_, err := ctgRepo.InsertCategory(ctx, nil, testCtg)
		require.NoError(t, err)

		ctgs, err := ctgRepo.GetCategories(ctx)
		require.NoError(t, err)

		foundCtg := ctgs[0]
		assert.Equal(t, testCtg.Name, foundCtg.Name)

		t.Cleanup(func() {
			cleanDB(t, ctx, "product_categories")
		})
	})
}

// TestCategoryRepository_UpdateCategory проверяет, что метод UpdateCategory корректно обновляет информацию о категории по её ID.
func TestCategoryRepository_UpdateCategory(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное обновление", func(t *testing.T) {
		err := seedTestData(ctx)
		require.NoError(t, err)

		rootID, err := ctgRepo.GetCategoryID(ctx, nil, "test_category")
		require.NoError(t, err)
		require.NotZero(t, rootID)

		testCtg := entities.ProductCategory{Name: "test", ParentID: int(rootID)}
		ctgID, err := ctgRepo.InsertCategory(ctx, nil, testCtg)
		require.NoError(t, err)
		require.NotZero(t, ctgID)

		updates := map[string]any{
			"name": "updated_name",
		}
		err = ctgRepo.UpdateCategory(ctx, ctgID, updates)
		require.NoError(t, err)

		category, err := ctgRepo.GetCategoryByID(ctx, ctgID)
		require.NoError(t, err, "GetCategories Error")
		require.NotNil(t, category)
		require.Equal(t, "updated_name", category.Name)

		t.Cleanup(func() {
			cleanDB(t, ctx, "product_categories")
		})
	})
}

// TestCategoryRepository_DeleteCategoryByID проверяет, что метод DeleteCategoryByID корректно удаляет категорию по её ID.
func TestCategoryRepository_DeleteCategoryByID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное удаление", func(t *testing.T) {
		err := seedTestData(ctx)
		require.NoError(t, err)

		rootID, err := ctgRepo.GetCategoryID(ctx, nil, "test_category")
		require.NoError(t, err)
		require.NotZero(t, rootID)

		testCtg = entities.ProductCategory{Name: "test", ParentID: int(rootID)}
		ctgID, err := ctgRepo.InsertCategory(ctx, nil, testCtg)
		require.NoError(t, err)
		require.NotZero(t, ctgID)

		err = ctgRepo.DeleteCategoryByID(ctx, ctgID)
		require.NoError(t, err)

		category, err := ctgRepo.GetCategoryByID(ctx, ctgID)
		require.Nil(t, category)
		require.Error(t, err, "GetCategories Error")

		t.Cleanup(func() {
			cleanDB(t, ctx, "product_categories")
		})
	})
}

// TestCategoryRepository_GetCategoryID проверяет, что метод GetCategoryID корректно возвращает ID категории по её имени.
func TestCategoryRepository_GetCategoryID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное удаление", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")

		ctgID, _ := ctgRepo.InsertCategory(ctx, nil, testCtg)
		require.NotZero(t, ctgID)

		getCtgID, err := ctgRepo.GetCategoryID(ctx, nil, testCtg.Name)
		require.NoError(t, err)
		require.Equal(t, ctgID, getCtgID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "product_categories")
		})
	})
}
