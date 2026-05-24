package handlers_test

import (
	"Brewery/internal/http/handlers"
	"Brewery/internal/http/handlers/mocks"
	"Brewery/internal/http/middleware"
	"Brewery/internal/http/routers"
	"Brewery/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gojuno/minimock/v3"
)

type testEnv struct {
	Router        *gin.Engine
	BeerMock      *mocks.BeerServiceMock
	EnumMock      *mocks.EnumServiceMock
	ParameterMock *mocks.ParametersServiceMock
	AggregateMock *mocks.AggregateServiceMock

	JWT string
}

// setupTestEnv устанавливает необходимые переменные окружения для тестов.
func setupTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret-key-for-jwt")
	t.Setenv("ADMIN_USERNAME", "admin")
	t.Setenv("ADMIN_PASSWORD", "admin")
}

// setupIntegrationRouter инициализирует тестовый сервер с моками и необходимыми middleware для интеграционных тестов.
func setupIntegrationRouter(beerServiceM *mocks.BeerServiceMock, enumServiceM *mocks.EnumServiceMock, parametersServiseM *mocks.ParametersServiceMock, aggregateServiceM *mocks.AggregateServiceMock) *gin.Engine {
	gin.SetMode(gin.TestMode)

	logCtx, err := logger.NewLoggerContext(context.Background(), true)
	if err != nil {
		panic(err)
	}

	log, ok := logger.GetLoggerFromCtx(logCtx)
	if !ok {
		panic("GetLoggerFromCtx")
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(func(c *gin.Context) {
		ctxWithLogger := logger.NewContextWithLogger(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctxWithLogger)
		c.Next()
	})
	engine.Use(middleware.RequestContextMiddleware())
	engine.Use(middleware.MetricsMiddleware())

	h := handlers.Handlers{
		CategoryHandler:   handlers.NewCategoriesHandlers(beerServiceM),
		BeersHandler:      handlers.NewBeersHandlers(beerServiceM),
		ReviewHandler:     handlers.NewReviewsHandlers(beerServiceM),
		EnumHandler:       handlers.NewEnumHandlers(enumServiceM),
		ParametersHandler: handlers.NewParametersHandlers(parametersServiseM),
		AggregatesHandler: handlers.NewAggregateHandlers(aggregateServiceM),
		AuthHandler:       handlers.NewAuthHandlers(),
	}
	routers.RegisterRoutes(engine, h)

	return engine
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	setupTestEnv(t)

	mc := minimock.NewController(t)
	beerServiceMock := mocks.NewBeerServiceMock(mc)
	enumServiceMock := mocks.NewEnumServiceMock(mc)
	parametersServiceMock := mocks.NewParametersServiceMock(mc)
	aggregateServiceMock := mocks.NewAggregateServiceMock(mc)

	router := setupIntegrationRouter(beerServiceMock, enumServiceMock, parametersServiceMock, aggregateServiceMock)

	body := strings.NewReader(`{"username":"admin", "password": "admin"}`)
	req := httptest.NewRequestWithContext(t.Context(), "POST", "/api/login", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		panic(err)
	}

	return &testEnv{
		Router:        router,
		BeerMock:      beerServiceMock,
		EnumMock:      enumServiceMock,
		ParameterMock: parametersServiceMock,
		AggregateMock: aggregateServiceMock,

		JWT: resp["token"],
	}
}

func (e *testEnv) DoRequest(ctx context.Context, jwt, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(ctx, method, path, body)
	if body == nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if jwt != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	}
	w := httptest.NewRecorder()
	e.Router.ServeHTTP(w, req)
	return w
}
