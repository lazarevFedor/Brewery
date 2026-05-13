package repository_test

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"Brewery/internal/repository/queries"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAggregateRepository_InsertGetApply проверяет вставку, извлечение и применение агрегата к категории и ее потомкам.
func TestAggregateRepository_InsertGetApply(t *testing.T) {
	ctx := t.Context()
	aggRepo := repository.NewAggregatePostgres(testDB)

	cleanDB(t, ctx, "aggregates")
	cleanDB(t, ctx, "numeric_parameters")
	cleanDB(t, ctx, "enum_parameters")

	rootID, err := ctgRepo.GetCategoryID(ctx, "test_category")
	require.NoError(t, err)
	require.NotZero(t, rootID)

	childID, err := ctgRepo.InsertCategory(ctx, entities.ProductCategory{Name: "agg_child", ParentID: int(rootID)})
	require.NoError(t, err)
	grandChildID, err := ctgRepo.InsertCategory(ctx, entities.ProductCategory{Name: "agg_grand_child", ParentID: int(childID)})
	require.NoError(t, err)

	nInherit := insertNumericParam(t, ctx, 1, 10, "n_inherit", true)
	nLocal := insertNumericParam(t, ctx, 11, 20, "n_local", false)
	eInherit := insertEnumParam(t, ctx, 101, true)
	eLocal := insertEnumParam(t, ctx, 202, false)

	agg := &entities.Aggregate{
		Name:              "agg_test",
		Description:       "agg desc",
		NumericParameters: []int{int(nInherit), int(nLocal)},
		EnumParameters:    []int{int(eInherit), int(eLocal)},
	}

	got, err := aggRepo.InsertAggregate(ctx, agg)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotZero(t, got.ID)

	list, err := aggRepo.GetAggregates(ctx, "agg_test")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)
	found := false
	for _, a := range list {
		if a.ID == got.ID {
			found = true
			assert.ElementsMatch(t, []int{int(nInherit), int(nLocal)}, a.NumericParameters)
			assert.ElementsMatch(t, []int{int(eInherit), int(eLocal)}, a.EnumParameters)
		}
	}
	assert.True(t, found)

	rows, err := aggRepo.ApplyAggregate(ctx, rootID, got.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, rows, 1)

	rootNums, rootEnums := fetchCategoryParams(t, ctx, rootID)
	childNums, childEnums := fetchCategoryParams(t, ctx, childID)
	grandNums, grandEnums := fetchCategoryParams(t, ctx, grandChildID)

	assert.ElementsMatch(t, []int{int(nInherit), int(nLocal)}, rootNums)
	assert.ElementsMatch(t, []int{int(eInherit), int(eLocal)}, rootEnums)
	assert.ElementsMatch(t, []int{int(nInherit)}, childNums)
	assert.ElementsMatch(t, []int{int(eInherit)}, childEnums)
	assert.ElementsMatch(t, []int{int(nInherit)}, grandNums)
	assert.ElementsMatch(t, []int{int(eInherit)}, grandEnums)

	count := countQueryRows(t, ctx, queries.GetAggregates("agg_test"))
	assert.GreaterOrEqual(t, count, 1)
}

// TestAggregateRepository_UpdateAndDelete проверяет обновление и удаление агрегата, а также его влияние на категории после применения.
func TestAggregateRepository_UpdateAndDelete(t *testing.T) {
	ctx := t.Context()
	aggRepo := repository.NewAggregatePostgres(testDB)

	cleanDB(t, ctx, "aggregates")
	cleanDB(t, ctx, "numeric_parameters")
	cleanDB(t, ctx, "enum_parameters")

	rootID, err := ctgRepo.GetCategoryID(ctx, "test_category")
	require.NoError(t, err)
	require.NotZero(t, rootID)

	childID, err := ctgRepo.InsertCategory(ctx, entities.ProductCategory{Name: "agg_upd_child", ParentID: int(rootID)})
	require.NoError(t, err)

	n1 := insertNumericParam(t, ctx, 1, 2, "u_n1", false)
	n2 := insertNumericParam(t, ctx, 3, 4, "u_n2", true)
	e1 := insertEnumParam(t, ctx, 10, false)

	agg := &entities.Aggregate{
		Name:              "agg_upd",
		Description:       "before",
		NumericParameters: []int{int(n1)},
		EnumParameters:    []int{int(e1)},
	}

	inserted, err := aggRepo.InsertAggregate(ctx, agg)
	require.NoError(t, err)
	require.NotNil(t, inserted)

	updates := map[string]any{
		"name":                  "agg_upd_new",
		"description":           "after",
		"numeric_parameter_ids": []uint{n2},
	}

	updated, err := aggRepo.UpdateAggregate(ctx, inserted.ID, updates)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "agg_upd_new", updated.Name)
	assert.Equal(t, "after", updated.Description)
	assert.ElementsMatch(t, []int{int(n2)}, updated.NumericParameters)

	rows, err := aggRepo.ApplyAggregate(ctx, rootID, updated.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, rows, 1)

	rootNums, _ := fetchCategoryParams(t, ctx, rootID)
	childNums, _ := fetchCategoryParams(t, ctx, childID)

	assert.Contains(t, rootNums, int(n2))
	assert.Contains(t, childNums, int(n2))

	deleted, err := aggRepo.DeleteAggregate(ctx, updated.ID)
	require.NoError(t, err)
	require.NotNil(t, deleted)
	assert.Equal(t, updated.ID, deleted.ID)

	list, err := aggRepo.GetAggregates(ctx, "agg_upd_new")
	require.NoError(t, err)
	assert.Empty(t, list)
}
