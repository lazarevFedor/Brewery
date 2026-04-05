package handlers_test

import (
	"Brewery/internal/entities"
	"Brewery/internal/http/handlers"
	"Brewery/internal/http/handlers/mocks"
	"Brewery/internal/http/middleware"
	"Brewery/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationRouter инициализирует тестовый сервер с моками и необходимыми middleware для интеграционных тестов.
func setupIntegrationRouter(t *testing.T, svc *mocks.BeerServiceMock) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	logCtx, err := logger.NewLoggerContext(context.Background(), false)
	require.NoError(t, err)

	log, ok := logger.GetLoggerFromCtx(logCtx)
	require.True(t, ok)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(func(c *gin.Context) {
		ctxWithLogger := logger.NewContextWithLogger(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctxWithLogger)
		c.Next()
	})
	engine.Use(middleware.RequestContextMiddleware())
	engine.Use(middleware.MetricsMiddleware())

	categoryHandler := handlers.NewCategoriesHandlers(svc)
	beersHandler := handlers.NewBeersHandlers(svc)

	engine.POST("/api/categories", categoryHandler.CreateCategory)
	engine.GET("/api/categories/:id", categoryHandler.GetCategoryById)
	engine.PATCH("/api/categories/:id", categoryHandler.UpdateCategory)
	engine.DELETE("/api/categories/:id", categoryHandler.DeleteCategory)
	engine.GET("/api/categories", categoryHandler.GetAllCategories)
	engine.GET("/api/categories/:id/beers", categoryHandler.GetBeersByCategory)
	engine.GET("/api/categories/parent/:id", categoryHandler.GetParentCategory)
	engine.GET("/api/categories/children/:id", categoryHandler.GetChildCategory)

	engine.POST("/api/beers", beersHandler.CreateBeer)
	engine.PATCH("/api/beers/:id", beersHandler.UpdateBeer)
	engine.DELETE("/api/beers/:id", beersHandler.DeleteBeer)
	engine.GET("/api/beers", beersHandler.GetAllBeers)
	engine.POST("/api/beers/reviews/:id", beersHandler.CreateBeerReview)

	return engine
}

// TestServer_GetCategoryByID_UsesPathParam проверяет, что при запросе категории по ID используется правильный путь и передается правильный ID в сервис.
func TestServer_GetCategoryByID_UsesPathParam(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.GetCategoryByIDMock.
		Expect(minimock.AnyContext, 123).
		Return(&entities.ProductCategory{ID: 123, Name: "lager"}, nil)

	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/categories/123", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestServer_CreateCategory_ReturnsCreated проверяет, что при создании новой категории возвращается статус 201 Created и что тело запроса корректно обрабатывается.
func TestServer_CreateCategory_ReturnsCreated(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	router := setupIntegrationRouter(t, serviceMock)

	body := strings.NewReader(`{"name":"lager","parent_id":1}`)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/categories", body)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusCreated, resp.Code)
}

// TestServer_UpdateCategory_UsesPathParam проверяет, что при обновлении категории используется правильный путь и передается правильный ID в сервис.
func TestServer_UpdateCategory_UsesPathParam(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.UpdateCategoryMock.
		Expect(minimock.AnyContext, uint(12), map[string]any{"name": "ale"}).
		Return(nil)

	router := setupIntegrationRouter(t, serviceMock)

	body := strings.NewReader(`{"name":"ale"}`)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/api/categories/12", body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestServer_DeleteCategory_UsesPathParam проверяет, что при удалении категории используется правильный путь и передается правильный ID в сервис.
func TestServer_DeleteCategory_UsesPathParam(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.DeleteCategoryMock.
		Expect(minimock.AnyContext, 13).
		Return(nil)

	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/categories/13", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestServer_GetAllCategories_ReturnsOK проверяет, что при запросе всех категорий возвращается статус 200 OK и что сервис вызывается с правильными параметрами.
func TestServer_GetAllCategories_ReturnsOK(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.GetAllCategoriesMock.
		Expect(minimock.AnyContext).
		Return([]entities.ProductCategory{{ID: 1, Name: "lager"}}, nil)

	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/categories", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestServer_GetBeersByCategory_NewRoute_WorksWithIDAndPagination проверяет, что при запросе пива по категории с использованием нового маршрута возвращается статус 200 OK и что параметры пагинации корректно обрабатываются.
func TestServer_GetBeersByCategory_NewRoute_WorksWithIDAndPagination(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.GetBeersByCategoryMock.
		Expect(minimock.AnyContext, uint(42), uint64(1), uint64(1)).
		Return([]entities.Beer{
			{Name: "one", Rating: 4.1},
			{Name: "two", Rating: 4.2},
			{Name: "three", Rating: 4.3},
		}, nil)

	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/categories/42/beers?offset=1&limit=1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	var got []entities.Beer
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	require.Len(t, got, 3)
	assert.Equal(t, "one", got[0].Name)
	assert.Equal(t, "three", got[2].Name)
}

// TestServer_GetBeersByCategory_InvalidID_ReturnsBadRequest проверяет, что при запросе пива по категории с нечисловым ID возвращается статус 400 Bad Request и корректное сообщение об ошибке.
func TestServer_GetBeersByCategory_InvalidID_ReturnsBadRequest(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/categories/not-a-number/beers", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)

	var got map[string]string
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	assert.Equal(t, "invalid id", got["error"])
}

// TestServer_GetParentCategory_UsesPathParam проверяет, что при запросе родительской категории используется правильный путь и передается правильный ID в сервис.
func TestServer_GetParentCategory_UsesPathParam(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.GetParentCategoryMock.
		Expect(minimock.AnyContext, 14).
		Return(&entities.ProductCategory{ID: 1, Name: "parent"}, nil)

	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/categories/parent/14", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestServer_GetChildCategory_UsesPathParam проверяет, что при запросе дочерней категории используется правильный путь и передается правильный ID в сервис.
func TestServer_GetChildCategory_UsesPathParam(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.GetChildCategoryMock.
		Expect(minimock.AnyContext, 15).
		Return(&entities.ProductCategory{ID: 2, Name: "child"}, nil)

	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/categories/children/15", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestServer_UpdateBeer_UsesPathParam проверяет, что при обновлении пива используется правильный путь и передается правильный ID в сервис.
func TestServer_UpdateBeer_UsesPathParam(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.UpdateBeerMock.
		Expect(minimock.AnyContext, uint(77), map[string]any{"name": "ipa"}).
		Return(uint(77), nil)

	router := setupIntegrationRouter(t, serviceMock)

	body := strings.NewReader(`{"name":"ipa"}`)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/api/beers/77", body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestServer_CreateBeer_ReturnsCreated проверяет, что при создании нового пива возвращается статус 201 Created и что тело запроса корректно обрабатывается.
func TestServer_CreateBeer_ReturnsCreated(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	router := setupIntegrationRouter(t, serviceMock)

	body := strings.NewReader(`{"name":"ipa"}`)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/beers", body)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusCreated, resp.Code)
}

// TestServer_DeleteBeer_UsesPathParam проверяет, что при удалении пива используется правильный путь и передается правильный ID в сервис.
func TestServer_DeleteBeer_UsesPathParam(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.DeleteBeerMock.
		Expect(minimock.AnyContext, 16).
		Return(nil)

	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/beers/16", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

// TestServer_GetAllBeers_ReturnsOK проверяет, что при запросе всех изделий пива возвращается статус 200 OK и что сервис вызывается с правильными параметрами пагинации.
func TestServer_GetAllBeers_ReturnsOK(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.GetAllBeersMock.
		Expect(minimock.AnyContext, uint64(10), uint64(0)).
		Return([]entities.Beer{{Name: "ipa", Rating: 4.5}}, nil)

	router := setupIntegrationRouter(t, serviceMock)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/beers?offset=0&limit=10", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	var got []entities.Beer
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "ipa", got[0].Name)
}

// TestServer_CreateBeerReview_UsesPathParam проверяет, что при создании отзыва на пиво используется правильный путь и передается правильный ID пива в сервис.
func TestServer_CreateBeerReview_UsesPathParam(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := mocks.NewBeerServiceMock(mc)
	serviceMock.CreateBeerReviewMock.Set(func(ctx context.Context, review *entities.Review) (uint, error) {
		require.Equal(t, uint(17), review.BeerID)
		require.Equal(t, "great", review.Body)
		require.InDelta(t, 4.5, float64(review.Rating), 1e-6)

		return 1, nil
	})

	router := setupIntegrationRouter(t, serviceMock)

	body := strings.NewReader(`{"body":"great","rating":4.5}`)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/beers/reviews/17", body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}
