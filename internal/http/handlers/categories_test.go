package handlers_test

import (
	"Brewery/internal/entities"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
)

// Test_GetCategoryByID_UsesPathParam проверяет, что при запросе категории по ID используется правильный путь и передается правильный ID в сервис.
func Test_GetCategoryByID_UsesPathParam(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.GetCategoryByIDMock.
		Expect(minimock.AnyContext, 123).
		Return(&entities.ProductCategory{ID: 123, Name: "lager"}, nil)

	resp := testEnv.DoRequest(ctx, "", http.MethodGet, "/api/categories/123", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}

// Test_CreateCategory_ReturnsCreated проверяет, что при создании новой категории возвращается статус 201 Created и что тело запроса корректно обрабатывается.
func Test_CreateCategory_ReturnsCreated(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.CreateCategoryMock.Set(func(ctx context.Context, ctg *entities.ProductCategory) (uint, error) {
		require.Equal(t, "lager", ctg.Name)
		require.Equal(t, 1, ctg.ParentID)

		return 1, nil
	})

	body := strings.NewReader(`{"name":"lager","parent_id":1}`)
	resp := testEnv.DoRequest(ctx, testEnv.JWT, http.MethodPost, "/api/categories", body)

	require.Equal(t, http.StatusCreated, resp.Code)
}

// Test_UpdateCategory_UsesPathParam проверяет, что при обновлении категории используется правильный путь и передается правильный ID в сервис.
func Test_UpdateCategory_UsesPathParam(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.UpdateCategoryMock.
		Expect(minimock.AnyContext, uint(12), map[string]any{"name": "ale"}).
		Return(nil)

	body := strings.NewReader(`{"name":"ale"}`)
	resp := testEnv.DoRequest(ctx, testEnv.JWT, http.MethodPatch, "/api/categories/12", body)

	require.Equal(t, http.StatusOK, resp.Code)
}

// Test_DeleteCategory_UsesPathParam проверяет, что при удалении категории используется правильный путь и передается правильный ID в сервис.
func Test_DeleteCategory_UsesPathParam(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.DeleteCategoryMock.
		Expect(minimock.AnyContext, 13).
		Return(nil)

	resp := testEnv.DoRequest(ctx, testEnv.JWT, http.MethodDelete, "/api/categories/13", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}

// Test_GetAllCategories_ReturnsOK проверяет, что при запросе всех категорий возвращается статус 200 OK и что сервис вызывается с правильными параметрами.
func Test_GetAllCategories_ReturnsOK(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.GetAllCategoriesMock.
		Expect(minimock.AnyContext).
		Return([]entities.ProductCategory{{ID: 1, Name: "lager"}}, nil)

	resp := testEnv.DoRequest(ctx, "", http.MethodGet, "/api/categories", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}

// Test_GetParentCategory_UsesPathParam проверяет, что при запросе родительской категории используется правильный путь и передается правильный ID в сервис.
func Test_GetParentCategory_UsesPathParam(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.GetParentCategoryMock.
		Expect(minimock.AnyContext, 14).
		Return(&entities.ProductCategory{ID: 1, Name: "parent"}, nil)

	resp := testEnv.DoRequest(ctx, "", http.MethodGet, "/api/categories/parent/14", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}

// Test_GetChildCategory_UsesPathParam проверяет, что при запросе дочерней категории используется правильный путь и передается правильный ID в сервис.
func Test_GetChildCategory_UsesPathParam(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.GetChildCategoriesMock.
		Expect(minimock.AnyContext, 15).
		Return([]entities.ProductCategory{{ID: 2, Name: "child"}}, nil)

	resp := testEnv.DoRequest(ctx, "", http.MethodGet, "/api/categories/children/15", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}
