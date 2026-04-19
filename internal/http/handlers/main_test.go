package handlers_test

import (
	"Brewery/internal/http/handlers"
	"Brewery/internal/http/handlers/mocks"
	"Brewery/internal/http/handlers/routers"
	"Brewery/internal/http/middleware"
	"Brewery/pkg/logger"
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gojuno/minimock/v3"
)

type testEnv struct {
	Router *gin.Engine
	Mock   *mocks.BeerServiceMock
}

// setupIntegrationRouter инициализирует тестовый сервер с моками и необходимыми middleware для интеграционных тестов.
func setupIntegrationRouter(svc *mocks.BeerServiceMock) *gin.Engine {

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

	categoryHandler := handlers.NewCategoriesHandlers(svc)
	beersHandler := handlers.NewBeersHandlers(svc)

	routers.RegisterRoutes(engine, categoryHandler, beersHandler)
	return engine
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	mc := minimock.NewController(t)
	serviceMock := mocks.NewBeerServiceMock(mc)

	router := setupIntegrationRouter(serviceMock)

	return &testEnv{
		Router: router,
		Mock:   serviceMock,
	}
}

func (e *testEnv) DoRequest(ctx context.Context, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(ctx, method, path, body)
	if body == nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	e.Router.ServeHTTP(w, req)
	return w
}
