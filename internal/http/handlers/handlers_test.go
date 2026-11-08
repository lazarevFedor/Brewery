package handlers_test

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Классы ---

func TestGetClass_InvalidID_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes/abc", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, apperrors.CodeInvalidID, errorBody(t, resp)["error"])
}

func TestGetClass_OK(t *testing.T) {
	env := newTestEnv(t)
	env.ClassMock.GetClassByIDMock.Expect(minimock.AnyContext, int64(5)).
		Return(&entities.Class{ID: 5, Name: "Светлое", ParentID: 1}, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes/5", nil)

	require.Equal(t, http.StatusOK, resp.Code)
	var got entities.Class
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	assert.Equal(t, "Светлое", got.Name)
}

func TestGetClass_AppErrorNotFound_Returns404(t *testing.T) {
	env := newTestEnv(t)
	env.ClassMock.GetClassByIDMock.Expect(minimock.AnyContext, int64(5)).
		Return(nil, apperrors.NotFound("Class with id=5 not found", errors.New("no rows")))

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes/5", nil)

	require.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, apperrors.CodeNotFound, errorBody(t, resp)["error"])
}

func TestGetClass_UnknownError_Returns500WithoutDetails(t *testing.T) {
	env := newTestEnv(t)
	env.ClassMock.GetClassByIDMock.Expect(minimock.AnyContext, int64(5)).
		Return(nil, errors.New("pq: connection refused"))

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes/5", nil)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
	body := errorBody(t, resp)
	assert.Equal(t, apperrors.CodeInternalError, body["error"])
	assert.NotContains(t, body["message"], "connection refused")
}

func TestGetClassParent_Root_Returns404(t *testing.T) {
	env := newTestEnv(t)
	env.ClassMock.GetParentClassMock.Expect(minimock.AnyContext, int64(1)).Return(nil, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes/parent/1", nil)

	require.Equal(t, http.StatusNotFound, resp.Code)
}

func TestListClasses_Empty_ReturnsEmptyArray(t *testing.T) {
	env := newTestEnv(t)
	env.ClassMock.GetAllClassesMock.Return(nil, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes", nil)

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `[]`, resp.Body.String())
}

func TestDeleteClass_Returns204(t *testing.T) {
	env := newTestEnv(t)
	env.ClassMock.DeleteClassByIDMock.Expect(minimock.AnyContext, int64(3)).Return(nil)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodDelete, "/api/classes/3", nil)

	require.Equal(t, http.StatusNoContent, resp.Code)
}

func TestCreateClass_EmptyName_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPost, "/api/classes", strings.NewReader(`{"parent_id":1}`))

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

// --- Изделия ---

func TestCreateProduct_MapsSwaggerFields(t *testing.T) {
	env := newTestEnv(t)
	env.ProductMock.CreateProductMock.Set(func(_ context.Context, p entities.Product) (*entities.Product, error) {
		assert.Equal(t, "Балтика 0", p.Name)
		assert.Equal(t, "Светлое", p.ProdClass.Name)
		assert.Equal(t, "L", p.Unit)
		assert.Equal(t, float32(100), p.Amount)

		p.ID = 7
		p.ProdClass.ID = 3
		return &p, nil
	})

	body := `{"name":"Балтика 0","class":{"name":"Светлое"},"amount":100,"units":"L"}`
	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPost, "/api/products", strings.NewReader(body))

	require.Equal(t, http.StatusCreated, resp.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	assert.Equal(t, "L", got["units"])
	assert.InDelta(t, 3, got["class_id"], 0)
	assert.Nil(t, got["base_id"])
}

func TestCreateProduct_WithoutClass_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPost, "/api/products", strings.NewReader(`{"name":"x"}`))

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestUpdateProduct_ConvertsIDsToInt64(t *testing.T) {
	env := newTestEnv(t)
	env.ProductMock.UpdateProductMock.Set(func(_ context.Context, id int64, updates map[string]any) (*entities.Product, error) {
		assert.Equal(t, int64(4), id)
		assert.Equal(t, int64(2), updates["class_id"])
		return &entities.Product{ID: id}, nil
	})

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPatch, "/api/products/4", strings.NewReader(`{"class_id":2}`))

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestUpdateProduct_FractionalID_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPatch, "/api/products/4", strings.NewReader(`{"base_id":1.5}`))

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestUpdateProduct_EmptyBody_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPatch, "/api/products/4", strings.NewReader(`{}`))

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListProducts_LimitIsClamped(t *testing.T) {
	env := newTestEnv(t)
	env.ProductMock.GetProductsMock.Expect(minimock.AnyContext, int32(100), int32(5)).Return(nil, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products?offset=5&limit=1000", nil)

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `[]`, resp.Body.String())
}

func TestListProducts_InvalidOffset_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products?offset=-1", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestSearchProducts_OnlyEqAllowed(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/search?filter=amount:gte:100", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestSearchProducts_Eq(t *testing.T) {
	env := newTestEnv(t)
	env.ProductMock.GetFilteredProductsMock.Set(func(_ context.Context, filters []entities.FilterParameter, classID *uint, _, _ int32) ([]entities.Product, error) {
		require.Len(t, filters, 1)
		assert.Equal(t, entities.FilterParameter{FieldName: "amount", Operation: entities.OpEqual, Value: 100}, filters[0])
		assert.Nil(t, classID)
		return nil, nil
	})

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/search?filter=amount:eq:100", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestSearchProductsByClass_RangeOperations(t *testing.T) {
	env := newTestEnv(t)
	env.ProductMock.GetFilteredProductsMock.Set(func(_ context.Context, filters []entities.FilterParameter, classID *uint, _, _ int32) ([]entities.Product, error) {
		require.Len(t, filters, 3)
		assert.Equal(t, entities.OpGreaterEqual, filters[0].Operation)
		assert.Equal(t, entities.OpLessEqual, filters[1].Operation)
		assert.Equal(t, entities.OpLess, filters[2].Operation)
		require.NotNil(t, classID)
		assert.Equal(t, uint(3), *classID)
		return nil, nil
	})

	resp := env.DoRequest(t.Context(), "", http.MethodGet,
		"/api/classes/3/products/search?filter=amount:gte:100&filter=amount:le:200&filter=abv:lt:5", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestSearchProductsByClass_InvalidField_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes/3/products/search?filter=amount%20or%201:eq:1", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListDerivedProducts_NoService_Returns501(t *testing.T) {
	env := newTestEnvWithOptions(t, envOptions{withoutDerived: true})

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/1/derived", nil)

	require.Equal(t, http.StatusNotImplemented, resp.Code)
}

// --- Материалы ---

func TestMaterials_NoService_Returns501(t *testing.T) {
	env := newTestEnvWithOptions(t, envOptions{withoutMaterials: true})

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/1/materials", nil)

	require.Equal(t, http.StatusNotImplemented, resp.Code)
}

func TestGetProductBomTree_OK(t *testing.T) {
	env := newTestEnv(t)
	env.MaterialMock.GetMaterialsByProductMock.Expect(minimock.AnyContext, int64(1)).Return([]entities.MaterialProduct{
		{
			Material: entities.Material{MaterialAmount: 12.5},
			Product:  entities.Product{ID: 5, Name: "Сусло"},
			Materials: []entities.MaterialProduct{
				{Material: entities.Material{MaterialAmount: 2}, Product: entities.Product{ID: 7, Name: "Солод"}},
			},
		},
	}, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/1/materials", nil)

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `[{"material_id":5,"name":"Сусло","material_amount":12.5,"materials":[
		{"material_id":7,"name":"Солод","material_amount":2,"materials":[]}]}]`, resp.Body.String())
}

func TestCalculateMaterialNorms_InvalidID_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/abc/materials/norms", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestCalculateMaterialNorms_OK(t *testing.T) {
	env := newTestEnv(t)
	env.MaterialMock.GetMaterialNormsMock.Expect(minimock.AnyContext, int64(1), float32(1)).
		Return([]entities.MaterialProduct{
			{
				Material: entities.Material{MaterialID: 5, MaterialAmount: 37.5},
				Product:  entities.Product{Name: "Хмель", Unit: "kg"},
			},
			{
				Material: entities.Material{MaterialAmount: 2},
				Product:  entities.Product{ID: 7, Name: "Солод", Unit: "kg"},
			},
		}, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/1/materials/norms", nil)

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `[
		{"material_id":5,"name":"Хмель","total_amount":37.5,"units":"kg"},
		{"material_id":7,"name":"Солод","total_amount":2,"units":"kg"}]`, resp.Body.String())
}

func TestCalculateMaterialNorms_PassesQuantity(t *testing.T) {
	env := newTestEnv(t)
	env.MaterialMock.GetMaterialNormsMock.Expect(minimock.AnyContext, int64(1), float32(2.5)).Return(nil, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/1/materials/norms?quantity=2.5", nil)

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `[]`, resp.Body.String())
}

func TestCalculateMaterialNorms_InvalidQuantity_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	for _, q := range []string{"0", "-1", "abc"} {
		resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/1/materials/norms?quantity="+q, nil)
		require.Equal(t, http.StatusBadRequest, resp.Code, "quantity=%s", q)
	}
}

func TestAddProductMaterial_SelfReference_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPost, "/api/products/1/materials",
		strings.NewReader(`{"material_id":1,"material_amount":2}`))

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAddProductMaterial_OK(t *testing.T) {
	env := newTestEnv(t)
	env.MaterialMock.CreateMaterialMock.
		Expect(minimock.AnyContext, entities.Material{MaterialID: 5, ProductID: 1, MaterialAmount: 2}).
		Return(&entities.Material{MaterialID: 5, ProductID: 1, MaterialAmount: 2}, nil)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPost, "/api/products/1/materials",
		strings.NewReader(`{"material_id":5,"material_amount":2}`))

	require.Equal(t, http.StatusCreated, resp.Code)
	assert.JSONEq(t, `{"material_id":5,"product_id":1,"material_amount":2}`, resp.Body.String())
}

func TestUpdateProductMaterial_OK(t *testing.T) {
	env := newTestEnv(t)
	env.MaterialMock.UpdateMaterialMock.Expect(minimock.AnyContext, int64(1), int64(5), float32(3)).
		Return(&entities.Material{MaterialID: 5, ProductID: 1, MaterialAmount: 3}, nil)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPatch, "/api/products/1/materials/5",
		strings.NewReader(`{"material_amount":3}`))

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `{"material_id":5,"product_id":1,"material_amount":3}`, resp.Body.String())
}

func TestDeleteProductMaterial_Returns204(t *testing.T) {
	env := newTestEnv(t)
	env.MaterialMock.DeleteMaterialMock.Expect(minimock.AnyContext, int64(1), int64(5)).Return(nil)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodDelete, "/api/products/1/materials/5", nil)

	require.Equal(t, http.StatusNoContent, resp.Code)
}

// --- История ---

func TestRegisterProductChange_EmptyChanges_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPost, "/api/products/1/history", strings.NewReader(`{"changes":"  "}`))

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestRegisterProductChange_OK(t *testing.T) {
	env := newTestEnv(t)
	createdAt := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	env.HistoryMock.CreateChangeRecordMock.
		Expect(minimock.AnyContext, &entities.HistoryRecord{ProductID: 1, Changes: "Норма хмеля увеличена"}).
		Return(&entities.HistoryRecord{ID: 10, ProductID: 1, Changes: "Норма хмеля увеличена", CreatedAt: createdAt}, nil)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPost, "/api/products/1/history",
		strings.NewReader(`{"changes":"Норма хмеля увеличена"}`))

	require.Equal(t, http.StatusCreated, resp.Code)
	assert.JSONEq(t, `{"id":10,"product_id":1,"changes":"Норма хмеля увеличена","created_at":"2026-11-07T12:00:00Z"}`, resp.Body.String())
}

func TestListProductHistory_InvalidDate_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/1/history?from=yesterday", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListProductHistory_PassesPeriod(t *testing.T) {
	env := newTestEnv(t)
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	env.HistoryMock.GetHistoryByProductMock.
		Expect(minimock.AnyContext, int64(1), from, time.Time{}, int32(10), int32(0)).
		Return(nil, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/products/1/history?from=2026-11-01T00:00:00Z", nil)

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `[]`, resp.Body.String())
}

func TestGetHistoryRecord_Empty_Returns404(t *testing.T) {
	env := newTestEnv(t)
	env.HistoryMock.GetHistoryRecordByIDMock.Expect(minimock.AnyContext, int64(9)).Return(nil, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/history/9", nil)

	require.Equal(t, http.StatusNotFound, resp.Code)
}

func TestHistory_NoService_Returns501(t *testing.T) {
	env := newTestEnvWithOptions(t, envOptions{withoutHistory: true})

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/history/1", nil)

	require.Equal(t, http.StatusNotImplemented, resp.Code)
}

// --- Перечисления, параметры, агрегаты ---

func TestGetEnum_RequiresFieldName(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodGet, "/api/enums", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestCreateEnum_ReturnsCreatedObject(t *testing.T) {
	env := newTestEnv(t)
	env.EnumMock.CreateEnumMock.Return(uint(4), nil)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPost, "/api/enums",
		strings.NewReader(`{"field_name":"city","enum_type":"string","is_active":true}`))

	require.Equal(t, http.StatusCreated, resp.Code)
	var got entities.EnumClass
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	assert.Equal(t, 4, got.ID)
	assert.Equal(t, "city", got.FieldName)
}

func TestListParameters_InvalidType_ReturnsBadRequest(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes/parameters?type=foo", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListParameters_RangeType(t *testing.T) {
	env := newTestEnv(t)
	env.ParameterMock.ListParametersMock.Expect(minimock.AnyContext, int64(2), entities.RangeParameterType).
		Return([]entities.RangeParameter{{ID: 1, FieldName: "abv"}}, nil, nil)

	resp := env.DoRequest(t.Context(), "", http.MethodGet, "/api/classes/parameters?class_id=2&type=range", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestExcludeParameters_OK(t *testing.T) {
	env := newTestEnv(t)
	env.ParameterMock.ExcludeFromClassMock.Expect(minimock.AnyContext, int64(2), []int{1}, []int(nil)).Return(1, nil)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPatch, "/api/classes/parameters/exclude/2",
		strings.NewReader(`{"range_param_ids":[1]}`))

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `{"message":"Parameters excluded successfully","params_excluded":1}`, resp.Body.String())
}

func TestApplyAggregate_RequiresAggregateID(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), env.JWT, http.MethodPatch, "/api/aggregates/apply/2", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestLogin_WrongPassword_Returns403(t *testing.T) {
	env := newTestEnv(t)

	resp := env.DoRequest(t.Context(), "", http.MethodPost, "/api/login", strings.NewReader(`{"username":"admin","password":"nope"}`))

	require.Equal(t, http.StatusForbidden, resp.Code)
	errorBody(t, resp)
}
