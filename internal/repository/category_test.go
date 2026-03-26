package repository_test

import (
	"Brewery/internal/entities"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryRepository_InsertCategory(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка", func(t *testing.T) {
		testCtg := entities.ProductCategory{Name: "test"}
		ctgID, err := ctgRepo.InsertCategory(ctx, testCtg)

		assert.NoError(t, err)
		assert.NotZero(t, ctgID)

		t.Cleanup(func() {
			cleanDB(t, ctx, ctgRepo.Pool, "product_categories")
		})
	})

	t.Run("Дублирование категории", func(t *testing.T) {
		// Заполняем первый раз
		testCtg := entities.ProductCategory{Name: "test"}
		ctgID, err := ctgRepo.InsertCategory(ctx, testCtg)

		assert.NoError(t, err)
		assert.NotZero(t, ctgID)
		

		// Заполняем второй раз
		_, err = ctgRepo.InsertCategory(ctx, testCtg)

		if err == nil {
			t.Error("Ожидалась ошибка уникальности, но запись создалась")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code != "23505" {
				t.Errorf("ожидался код ошибки 23505, получили %s", pgErr.Code)
			}
		} else {
			t.Errorf("ожидалась ошибка pgconn.PgError, получили %T", err)
		}

		t.Cleanup(func() {
			cleanDB(t, ctx, ctgRepo.Pool, "product_categories")
		})
	})

}

func TestCategoryRepository_GetCategories(t *testing.T){
	ctx := t.Context()

	t.Run("Пустая БД", func(t *testing.T){
		ctgs, err := ctgRepo.GetCategories(ctx)
		assert.Empty(t, ctgs)
		assert.NoError(t, err)
	})

	t.Run("Успешное нахождение 1 категории", func(t *testing.T){
		testCtg := entities.ProductCategory{Name: "test"}
		_, err := ctgRepo.InsertCategory(ctx, testCtg)
		require.NoError(t, err)

		ctgs, err := ctgRepo.GetCategories(ctx)
		require.NoError(t, err)

		foundCtg := ctgs[0]
		assert.Equal(t, testCtg.Name, foundCtg.Name)
	})
}
