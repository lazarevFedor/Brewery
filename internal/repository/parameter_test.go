// Package repository_test содержит тесты для слоя repository
package repository_test

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"Brewery/internal/repository/queries"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fetchCategoryParams - вспомогательная функция для получения связанных с категорией параметров из базы данных для проверки результатов тестов.
func fetchCategoryParams(t *testing.T, ctx context.Context, categoryID uint) ([]int, []int) {
	t.Helper()

	var numericIDs []int
	var enumIDs []int
	query := queries.SelectParameterIDsByCategory(categoryID, entities.MissingType)
	sql, args, err := query.ToSql()
	require.NoError(t, err)
	err = testDB.QueryRow(ctx, sql, args...).Scan(&numericIDs, &enumIDs)
	require.NoError(t, err)

	return numericIDs, enumIDs
}

// countQueryRows - вспомогательная функция для подсчета количества строк, возвращаемых запросом, для проверки результатов тестов.
func countQueryRows(t *testing.T, ctx context.Context, queryBuilder interface{ ToSql() (string, []any, error) }) int {
	t.Helper()

	sql, args, err := queryBuilder.ToSql()
	require.NoError(t, err)

	rows, err := testDB.Query(ctx, sql, args...)
	require.NoError(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	require.NoError(t, rows.Err())
	return count
}

// insertCategoryTree - вспомогательная функция для создания иерархии категорий в базе данных для проверки наследования параметров в тестах.
func insertCategoryTree(t *testing.T, ctx context.Context) (uint, uint, uint) {
	t.Helper()

	repo := repository.NewCategoryPostgres(testDB)

	rootID, err := repo.InsertCategory(ctx, entities.ProductCategory{Name: "param_root"})
	require.NoError(t, err)

	childID, err := repo.InsertCategory(ctx, entities.ProductCategory{Name: "param_child", ParentID: int(rootID)})
	require.NoError(t, err)

	grandChildID, err := repo.InsertCategory(ctx, entities.ProductCategory{Name: "param_grand_child", ParentID: int(childID)})
	require.NoError(t, err)

	return rootID, childID, grandChildID
}

// insertNumericParam - вспомогательная функция для вставки числового параметра в базу данных и возврата его ID для использования в тестах.
func insertNumericParam(t *testing.T, ctx context.Context, minVal, maxVal int, field string, inheritable bool) uint {
	t.Helper()

	repo := repository.NewParameterPostgres(testDB)
	got, err := repo.InsertNumericParameter(ctx, &entities.NumericParameter{
		MinValue:    minVal,
		MaxValue:    maxVal,
		FieldName:   field,
		EntityName:  "beer",
		Inheritable: inheritable,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	return got.ID
}

// insertEnumParam - вспомогательная функция для вставки перечислимого параметра в базу данных и возврата его ID для использования в тестах.
func insertEnumParam(t *testing.T, ctx context.Context, enumClassID uint, inheritable bool) uint {
	t.Helper()

	repo := repository.NewParameterPostgres(testDB)
	got, err := repo.InsertEnumParameter(ctx, &entities.EnumParameter{
		EnumClassID: enumClassID,
		Inheritable: inheritable,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	return got.ID
}

// TestParameterRepository_InsertNumericParameter тестирует вставку числового параметра
func TestParameterRepository_InsertNumericParameter(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Успешная вставка", func(t *testing.T) {
		cleanDB(t, ctx, "numeric_parameters")

		param := &entities.NumericParameter{
			MinValue:    1,
			MaxValue:    10,
			FieldName:   "amount",
			EntityName:  "beer",
			Inheritable: true,
		}

		got, err := repo.InsertNumericParameter(ctx, param)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.NotZero(t, got.ID)
		assert.Equal(t, param.MinValue, got.MinValue)
		assert.Equal(t, param.MaxValue, got.MaxValue)
		assert.Equal(t, param.FieldName, got.FieldName)
		assert.Equal(t, param.EntityName, got.EntityName)
		assert.Equal(t, param.Inheritable, got.Inheritable)

		count := countQueryRows(t, ctx, queries.SelectNumericParameters([]uint{got.ID}))
		assert.Equal(t, 1, count)
	})

	t.Run("Множественные вставки уникальны", func(t *testing.T) {
		cleanDB(t, ctx, "numeric_parameters")

		cases := []struct {
			name string
			in   *entities.NumericParameter
		}{
			{name: "first", in: &entities.NumericParameter{MinValue: 0, MaxValue: 5, FieldName: "a", EntityName: "e1", Inheritable: false}},
			{name: "second", in: &entities.NumericParameter{MinValue: 10, MaxValue: 20, FieldName: "b", EntityName: "e2", Inheritable: true}},
		}

		ids := make([]uint, 0, len(cases))
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := repo.InsertNumericParameter(ctx, tc.in)
				require.NoError(t, err)
				ids = append(ids, got.ID)
			})
		}

		assert.Len(t, ids, 2)
		assert.NotEqual(t, ids[0], ids[1])
	})
}

// TestParameterRepository_UpdateNumericParameter тестирует обновление числового параметра.
func TestParameterRepository_UpdateNumericParameter(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Обновление существующего параметра", func(t *testing.T) {
		cleanDB(t, ctx, "numeric_parameters")

		paramID := insertNumericParam(t, ctx, 1, 2, "f", false)

		updated, err := repo.UpdateNumericParameter(ctx, paramID, map[string]any{"min_val": 5, "inheritable": true})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, paramID, updated.ID)
		assert.Equal(t, 5, updated.MinValue)
		assert.True(t, updated.Inheritable)
	})

	t.Run("Пустое обновление возвращает ошибку", func(t *testing.T) {
		cleanDB(t, ctx, "numeric_parameters")
		paramID := insertNumericParam(t, ctx, 1, 2, "f", false)
		_, err := repo.UpdateNumericParameter(ctx, paramID, map[string]any{})
		require.Error(t, err)
	})

	t.Run("Несуществующий ID возвращает ошибку", func(t *testing.T) {
		cleanDB(t, ctx, "numeric_parameters")
		_, err := repo.UpdateNumericParameter(ctx, 99999, map[string]any{"min_val": 3})
		require.Error(t, err)
	})
}

// TestParameterRepository_ApplyParametersAndInheritance тестирует применение параметров к категории.
func TestParameterRepository_ApplyParametersAndInheritance(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Применение к корню распространяет только наследуемые параметры на потомков", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, grandChildID := insertCategoryTree(t, ctx)

		nInheritable := insertNumericParam(t, ctx, 1, 10, "inherited_numeric", true)
		nLocal := insertNumericParam(t, ctx, 11, 20, "local_numeric", false)
		eInheritable := insertEnumParam(t, ctx, 100, true)
		eLocal := insertEnumParam(t, ctx, 200, false)

		rowsAffected, err := repo.ApplyParameters(ctx, rootID, []int{int(nInheritable), int(nLocal), int(nInheritable)}, []int{int(eInheritable), int(eLocal), int(eInheritable)})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, rowsAffected, 1)

		rootNumeric, rootEnum := fetchCategoryParams(t, ctx, rootID)
		childNumeric, childEnum := fetchCategoryParams(t, ctx, childID)
		grandNumeric, grandEnum := fetchCategoryParams(t, ctx, grandChildID)

		assert.ElementsMatch(t, []int{int(nInheritable), int(nLocal)}, rootNumeric)
		assert.ElementsMatch(t, []int{int(eInheritable), int(eLocal)}, rootEnum)
		assert.ElementsMatch(t, []int{int(nInheritable)}, childNumeric)
		assert.ElementsMatch(t, []int{int(eInheritable)}, childEnum)
		assert.ElementsMatch(t, []int{int(nInheritable)}, grandNumeric)
		assert.ElementsMatch(t, []int{int(eInheritable)}, grandEnum)
	})

	t.Run("Пустой ввод при применении не изменяет массивы категорий", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, _ := insertCategoryTree(t, ctx)

		rowsAffected, err := repo.ApplyParameters(ctx, rootID, nil, nil)
		require.NoError(t, err)
		assert.Zero(t, rowsAffected)

		rootNumeric, rootEnum := fetchCategoryParams(t, ctx, rootID)
		childNumeric, childEnum := fetchCategoryParams(t, ctx, childID)
		assert.Empty(t, rootNumeric)
		assert.Empty(t, rootEnum)
		assert.Empty(t, childNumeric)
		assert.Empty(t, childEnum)
	})
}

// TestParameterRepository_DeleteNumericParameter тестирует удаление числового параметра
func TestParameterRepository_DeleteNumericParameter(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Триггер очистки удаляет числовой параметр из всех категорий", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, grandChildID := insertCategoryTree(t, ctx)
		paramID := insertNumericParam(t, ctx, 1, 10, "cleanup_numeric", true)

		_, err := repo.ApplyParameters(ctx, rootID, []int{int(paramID)}, nil)
		require.NoError(t, err)

		deleted, err := repo.DeleteNumericParameter(ctx, paramID)
		require.NoError(t, err)
		require.NotNil(t, deleted)
		assert.Equal(t, paramID, deleted.ID)

		for _, categoryID := range []uint{rootID, childID, grandChildID} {
			numericIDs, _ := fetchCategoryParams(t, ctx, categoryID)
			assert.NotContains(t, numericIDs, int(paramID))
		}

		count := countQueryRows(t, ctx, queries.SelectNumericParameters([]uint{paramID}))
		assert.Zero(t, count)
	})

	t.Run("Удаление несуществующего числового параметра возвращает ошибку", func(t *testing.T) {
		cleanDB(t, ctx, "numeric_parameters")
		_, err := repo.DeleteNumericParameter(ctx, 99999)
		require.Error(t, err)
	})
}

// TestParameterRepository_DeleteEnumParameter тестирует удаление перечислимого параметра
func TestParameterRepository_DeleteEnumParameter(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Триггер очистки удаляет перечислимый параметр из всех категорий", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, grandChildID := insertCategoryTree(t, ctx)
		paramID := insertEnumParam(t, ctx, 100, true)

		_, err := repo.ApplyParameters(ctx, rootID, nil, []int{int(paramID)})
		require.NoError(t, err)

		deleted, err := repo.DeleteEnumParameter(ctx, paramID)
		require.NoError(t, err)
		require.NotNil(t, deleted)
		assert.Equal(t, paramID, deleted.ID)

		for _, categoryID := range []uint{rootID, childID, grandChildID} {
			_, enumIDs := fetchCategoryParams(t, ctx, categoryID)
			assert.NotContains(t, enumIDs, int(paramID))
		}

		count := countQueryRows(t, ctx, queries.SelectEnumParameters([]uint{paramID}))
		assert.Zero(t, count)
	})

	t.Run("Удаление несуществующего перечислимого параметра возвращает ошибку", func(t *testing.T) {
		cleanDB(t, ctx, "enum_parameters")
		_, err := repo.DeleteEnumParameter(ctx, 99999)
		require.Error(t, err)
	})
}

// TestParameterRepository_UpdateInheritableAndInherit тестирует обновление флага наследования для числового и перечислимого параметра
func TestParameterRepository_UpdateInheritableAndInherit(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Установка inheritable в true распространяет числовой параметр на потомков", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, grandChildID := insertCategoryTree(t, ctx)

		nLocal := insertNumericParam(t, ctx, 11, 20, "local_numeric", false)

		_, err := repo.ApplyParameters(ctx, rootID, []int{int(nLocal)}, nil)
		require.NoError(t, err)

		rootNumeric, _ := fetchCategoryParams(t, ctx, rootID)
		childNumeric, _ := fetchCategoryParams(t, ctx, childID)
		grandNumeric, _ := fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootNumeric, int(nLocal))
		assert.NotContains(t, childNumeric, int(nLocal))
		assert.NotContains(t, grandNumeric, int(nLocal))

		updated, err := repo.UpdateNumericParameter(ctx, nLocal, map[string]any{"inheritable": true})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.True(t, updated.Inheritable)

		rootNumeric, _ = fetchCategoryParams(t, ctx, rootID)
		childNumeric, _ = fetchCategoryParams(t, ctx, childID)
		grandNumeric, _ = fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootNumeric, int(nLocal))
		assert.Contains(t, childNumeric, int(nLocal))
		assert.Contains(t, grandNumeric, int(nLocal))
	})

	t.Run("Установка inheritable в true распространяет перечислимый параметр на потомков", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, grandChildID := insertCategoryTree(t, ctx)

		eLocal := insertEnumParam(t, ctx, 200, false)

		_, err := repo.ApplyParameters(ctx, rootID, nil, []int{int(eLocal)})
		require.NoError(t, err)

		_, rootEnum := fetchCategoryParams(t, ctx, rootID)
		_, childEnum := fetchCategoryParams(t, ctx, childID)
		_, grandEnum := fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootEnum, int(eLocal))
		assert.NotContains(t, childEnum, int(eLocal))
		assert.NotContains(t, grandEnum, int(eLocal))

		updated, err := repo.UpdateEnumParameter(ctx, eLocal, map[string]any{"inheritable": true})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.True(t, updated.Inheritable)

		_, rootEnum = fetchCategoryParams(t, ctx, rootID)
		_, childEnum = fetchCategoryParams(t, ctx, childID)
		_, grandEnum = fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootEnum, int(eLocal))
		assert.Contains(t, childEnum, int(eLocal))
		assert.Contains(t, grandEnum, int(eLocal))
	})
}

// TestParameterRepository_UpdateNumericParameterInheritableFalsToTrue тестирует изменение флага наследования для числового параметра с false на true
func TestParameterRepository_UpdateNumericParameterInheritableFalsToTrue(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Переключение inheritable с false на true добавляет параметр потомкам", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, grandChildID := insertCategoryTree(t, ctx)

		nLocal := insertNumericParam(t, ctx, 5, 50, "toggle_numeric", false)

		_, err := repo.ApplyParameters(ctx, rootID, []int{int(nLocal)}, nil)
		require.NoError(t, err)

		rootNumeric, _ := fetchCategoryParams(t, ctx, rootID)
		childNumeric, _ := fetchCategoryParams(t, ctx, childID)
		grandNumeric, _ := fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootNumeric, int(nLocal))
		assert.NotContains(t, childNumeric, int(nLocal))
		assert.NotContains(t, grandNumeric, int(nLocal))

		updated, err := repo.UpdateNumericParameter(ctx, nLocal, map[string]any{"inheritable": true})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.True(t, updated.Inheritable)

		rootNumeric, _ = fetchCategoryParams(t, ctx, rootID)
		childNumeric, _ = fetchCategoryParams(t, ctx, childID)
		grandNumeric, _ = fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootNumeric, int(nLocal))
		assert.Contains(t, childNumeric, int(nLocal))
		assert.Contains(t, grandNumeric, int(nLocal))
	})
}

// TestParameterRepository_UpdateNumericParameterInheritableTrueToFalse тестирует изменение флага наследования для числового параметра с true на false
func TestParameterRepository_UpdateNumericParameterInheritableTrueToFalse(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Переключение inheritable с true на false удаляет параметр у потомков", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, grandChildID := insertCategoryTree(t, ctx)

		nInheritable := insertNumericParam(t, ctx, 10, 100, "toggle_numeric_inherit", true)

		_, err := repo.ApplyParameters(ctx, rootID, []int{int(nInheritable)}, nil)
		require.NoError(t, err)

		rootNumeric, _ := fetchCategoryParams(t, ctx, rootID)
		childNumeric, _ := fetchCategoryParams(t, ctx, childID)
		grandNumeric, _ := fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootNumeric, int(nInheritable))
		assert.Contains(t, childNumeric, int(nInheritable))
		assert.Contains(t, grandNumeric, int(nInheritable))

		updated, err := repo.UpdateNumericParameter(ctx, nInheritable, map[string]any{"inheritable": false})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.False(t, updated.Inheritable)

		rootNumeric, _ = fetchCategoryParams(t, ctx, rootID)
		childNumeric, _ = fetchCategoryParams(t, ctx, childID)
		grandNumeric, _ = fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootNumeric, int(nInheritable))
		assert.NotContains(t, childNumeric, int(nInheritable))
		assert.NotContains(t, grandNumeric, int(nInheritable))
	})
}

// TestParameterRepository_UpdateEnumParameterInheritableTrueToFalse тестирует изменение флага наследования для перечислимого параметра с true на false
func TestParameterRepository_UpdateEnumParameterInheritableTrueToFalse(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("Переключение inheritable с true на false удаляет перечислимый параметр у потомков", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, grandChildID := insertCategoryTree(t, ctx)

		eInheritable := insertEnumParam(t, ctx, 300, true)

		_, err := repo.ApplyParameters(ctx, rootID, nil, []int{int(eInheritable)})
		require.NoError(t, err)

		_, rootEnum := fetchCategoryParams(t, ctx, rootID)
		_, childEnum := fetchCategoryParams(t, ctx, childID)
		_, grandEnum := fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootEnum, int(eInheritable))
		assert.Contains(t, childEnum, int(eInheritable))
		assert.Contains(t, grandEnum, int(eInheritable))

		updated, err := repo.UpdateEnumParameter(ctx, eInheritable, map[string]any{"inheritable": false})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.False(t, updated.Inheritable)

		_, rootEnum = fetchCategoryParams(t, ctx, rootID)
		_, childEnum = fetchCategoryParams(t, ctx, childID)
		_, grandEnum = fetchCategoryParams(t, ctx, grandChildID)

		assert.Contains(t, rootEnum, int(eInheritable))
		assert.NotContains(t, childEnum, int(eInheritable))
		assert.NotContains(t, grandEnum, int(eInheritable))
	})
}

// TestParameterRepository_EdgeCases тестирует крайние случаи для репозитория параметров, включая получение параметров для категории без параметров.
func TestParameterRepository_EdgeCases(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewParameterPostgres(testDB)

	t.Run("GetParameters возвращает пустой набор для категории без параметров", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, _, _ := insertCategoryTree(t, ctx)

		n, e, err := repo.GetParameters(ctx, rootID, entities.MissingType)
		require.NoError(t, err)
		assert.Empty(t, n)
		assert.Empty(t, e)
	})

	t.Run("ApplyParameters удаляет дубликаты входных ID", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		rootID, childID, _ := insertCategoryTree(t, ctx)
		pid := insertNumericParam(t, ctx, 1, 2, "dup_test", false)

		rows, err := repo.ApplyParameters(ctx, rootID, []int{int(pid), int(pid), int(pid)}, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, rows, 1)

		rootNums, _ := fetchCategoryParams(t, ctx, rootID)
		childNums, _ := fetchCategoryParams(t, ctx, childID)

		assert.ElementsMatch(t, []int{int(pid)}, rootNums)
		assert.Empty(t, childNums)
	})

	t.Run("Переключение inheritable для непримененного параметра — нет действий", func(t *testing.T) {
		cleanDB(t, ctx, "product_categories")
		cleanDB(t, ctx, "numeric_parameters")
		cleanDB(t, ctx, "enum_parameters")

		pid := insertNumericParam(t, ctx, 1, 2, "orphan", false)

		updated, err := repo.UpdateNumericParameter(ctx, pid, map[string]any{"inheritable": true})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.True(t, updated.Inheritable)
	})
}
