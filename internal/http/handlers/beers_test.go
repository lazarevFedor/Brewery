package handlers_test

import (
	"Brewery/internal/entities"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetBeersByCategory_NewRoute_WorksWithIDAndPagination проверяет, что при запросе пива по категории с использованием нового маршрута возвращается статус 200 OK и что параметры пагинации корректно обрабатываются.
func TestGetBeersByCategory_NewRoute_WorksWithIDAndPagination(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.GetBeersByCategoryMock.
		Expect(minimock.AnyContext, uint(42), uint64(1), uint64(1)).
		Return([]entities.Beer{
			{Name: "one", Rating: 4.1},
			{Name: "two", Rating: 4.2},
			{Name: "three", Rating: 4.3},
		}, nil)

	resp := testEnv.DoRequest(ctx, http.MethodGet, "/api/categories/beers/42?offset=1&limit=1", nil)

	require.Equal(t, http.StatusOK, resp.Code)

	var got []entities.Beer
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	require.Len(t, got, 3)
	assert.Equal(t, "one", got[0].Name)
	assert.Equal(t, "three", got[2].Name)
}

// TestGetBeersByCategory_InvalidID_ReturnsBadRequest проверяет, что при запросе пива по категории с нечисловым ID возвращается статус 400 Bad Request и корректное сообщение об ошибке.
func TestGetBeersByCategory_InvalidID_ReturnsBadRequest(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)

	resp := testEnv.DoRequest(ctx, http.MethodGet, "/api/categories/beers/not-a-number", nil)

	require.Equal(t, http.StatusBadRequest, resp.Code)

	var got map[string]string
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	assert.Equal(t, "invalid id", got["error"])
}

// TestUpdateBeer_UsesPathParam проверяет, что при обновлении пива используется правильный путь и передается правильный ID в сервис.
func TestUpdateBeer_UsesPathParam(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.UpdateBeerMock.
		Expect(minimock.AnyContext, uint(77), map[string]any{"name": "ipa"}).
		Return(uint(77), nil)

	body := strings.NewReader(`{"name":"ipa"}`)
	resp := testEnv.DoRequest(ctx, http.MethodPatch, "/api/beers/77", body)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestCreateBeer_ReturnsCreated проверяет, что при создании нового пива возвращается статус 201 Created и что тело запроса корректно обрабатывается.
func TestCreateBeer_ReturnsCreated(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.CreateBeerMock.Set(
		func(ctx context.Context, beer *entities.Beer) (uint, error) {
			require.Equal(t, "ipa", beer.Name)

			return 1, nil
		})

	body := strings.NewReader(`{"name":"ipa"}`)
	resp := testEnv.DoRequest(ctx, http.MethodPost, "/api/beers", body)

	require.Equal(t, http.StatusCreated, resp.Code)
}

// TestDeleteBeer_UsesPathParam проверяет, что при удалении пива используется правильный путь и передается правильный ID в сервис.
func TestDeleteBeer_UsesPathParam(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.DeleteBeerMock.
		Expect(minimock.AnyContext, 16).
		Return(nil)

	resp := testEnv.DoRequest(ctx, http.MethodDelete, "/api/beers/16", nil)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestGetAllBeers_ReturnsOK проверяет, что при запросе всех изделий пива возвращается статус 200 OK и что сервис вызывается с правильными параметрами пагинации.
func TestGetAllBeers_ReturnsOK(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.GetAllBeersMock.
		Expect(minimock.AnyContext, uint64(10), uint64(0)).
		Return([]entities.Beer{{Name: "ipa", Rating: 4.5}}, nil)

	resp := testEnv.DoRequest(ctx, http.MethodGet, "/api/beers?offset=0&limit=10", nil)

	require.Equal(t, http.StatusOK, resp.Code)

	var got []entities.Beer
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "ipa", got[0].Name)
}

// TestCreateBeerReview_UsesPathParam проверяет, что при создании отзыва на пиво используется правильный путь и передается правильный ID пива в сервис.
func TestCreateBeerReview_UsesPathParam(t *testing.T) {
	ctx := t.Context()
	testEnv := newTestEnv(t)
	serviceMock := testEnv.BeerMock

	serviceMock.CreateReviewMock.Set(func(ctx context.Context, review *entities.Review) (uint, error) {
		require.Equal(t, uint(17), review.BeerID)
		require.Equal(t, "great", review.Body)
		require.InDelta(t, 4.5, float64(review.Rating), 1e-6)

		return 1, nil
	})

	body := strings.NewReader(`{"body":"great","rating":4.5}`)
	resp := testEnv.DoRequest(ctx, http.MethodPost, "/api/reviews/17", body)

	require.Equal(t, http.StatusOK, resp.Code)
}

// type args struct {
// 	beerID uint
// 	offset uint
// 	limit  uint
// }
// func TestGetBeersByCategory(t *testing.T) {
// 	ctx := t.Context()
// 	testEnv := newTestEnv(t)
// 	serviceMock := testEnv.BeerMock

// 	tests := []struct {
// 		name       string
// 		args       args
// 		setup      func(*mocks.BeerServiceMock)
// 		wantStatus int
// 		wantResult []entities.Beer
// 	}{
// 		{
// 			name: "success Get",
// 			args: args{
// 				beerID: 42,
// 				limit:  1,
// 				offset: 1,
// 			},
// 			setup: func(m *mocks.BeerServiceMock){
// 				m.GetBeersByCategoryMock.
// 				Expect(minimock.AnyContext, 42, 1, 1).
// 				Return([]entities.Beer{
// 					{Name: "one", Rating: 4.1},
// 					{Name: "two", Rating: 4.2},
// 					{Name: "three", Rating: 4.3},
// 				}, nil),
// 			},
// 			wantStatus
// 		},
// 	}

// 	for _, tt := range tests {
// 		var got []entities.Beer
// 		resp := testEnv.DoRequest(ctx, http.MethodGet, tt.url, nil)
// 		err := json.Unmarshal(resp.Body.Bytes(), &got)

// 		require.NoError(t, err)
// 		require.Len(t, got, 3)
// 		assert.Equal(t, "tt.mock.", got[0].Name)
// 		assert.Equal(t, "three", got[2].Name)
// 	}
// }
