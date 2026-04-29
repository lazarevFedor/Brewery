package handlers_test

import (
	"Brewery/internal/http/handlers"
	"Brewery/internal/http/handlers/mocks"
	"Brewery/internal/http/middleware"
	"Brewery/internal/http/routers"
	"Brewery/pkg/logger"
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gojuno/minimock/v3"
)

type testEnv struct {
	Router   *gin.Engine
	BeerMock *mocks.BeerServiceMock
	EnumMock *mocks.EnumServiceMock
}

// setupIntegrationRouter инициализирует тестовый сервер с моками и необходимыми middleware для интеграционных тестов.
func setupIntegrationRouter(beerServiceM *mocks.BeerServiceMock, enumServiceM *mocks.EnumServiceMock) *gin.Engine {
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
		CategoryHandler: handlers.NewCategoriesHandlers(beerServiceM),
		BeersHandler:    handlers.NewBeersHandlers(beerServiceM),
		ReviewHandler:   handlers.NewReviewsHandlers(beerServiceM),
		EnumHandler:     handlers.NewEnumHandlers(enumServiceM),
	}
	routers.RegisterRoutes(engine, h)

	return engine
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	mc := minimock.NewController(t)
	beerServiceMock := mocks.NewBeerServiceMock(mc)
	enumServiceMock := mocks.NewEnumServiceMock(mc)

	router := setupIntegrationRouter(beerServiceMock, enumServiceMock)

	return &testEnv{
		Router:   router,
		BeerMock: beerServiceMock,
		EnumMock: enumServiceMock,
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
