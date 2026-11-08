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
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	Router        *gin.Engine
	ClassMock     *mocks.ClassServiceMock
	ProductMock   *mocks.ProductServiceMock
	DerivedMock   *mocks.DerivedProductServiceMock
	MaterialMock  *mocks.MaterialServiceMock
	HistoryMock   *mocks.HistoryServiceMock
	EnumMock      *mocks.EnumServiceMock
	ParameterMock *mocks.ParametersServiceMock
	AggregateMock *mocks.AggregateServiceMock

	JWT string
}

// envOptions позволяет создать окружение без части сервисов, чтобы проверить ответ 501.
type envOptions struct {
	withoutMaterials bool
	withoutHistory   bool
	withoutDerived   bool
}

// setupTestEnv устанавливает необходимые переменные окружения для тестов.
func setupTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret-key-for-jwt")
	t.Setenv("ADMIN_USERNAME", "admin")
	t.Setenv("ADMIN_PASSWORD", "admin")
}

// setupRouter собирает роутер с теми же middleware, что и в main.
func setupRouter(t *testing.T, h handlers.Handlers) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	logCtx, err := logger.NewLoggerContext(context.Background(), true)
	require.NoError(t, err)

	log, ok := logger.GetLoggerFromCtx(logCtx)
	require.True(t, ok)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(logger.NewContextWithLogger(c.Request.Context(), log))
		c.Next()
	})
	engine.Use(middleware.RequestContextMiddleware())
	engine.Use(middleware.MetricsMiddleware())

	routers.RegisterRouters(engine, h)

	return engine
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	return newTestEnvWithOptions(t, envOptions{})
}

func newTestEnvWithOptions(t *testing.T, opts envOptions) *testEnv {
	t.Helper()
	setupTestEnv(t)

	mc := minimock.NewController(t)
	env := &testEnv{
		ClassMock:     mocks.NewClassServiceMock(mc),
		ProductMock:   mocks.NewProductServiceMock(mc),
		DerivedMock:   mocks.NewDerivedProductServiceMock(mc),
		MaterialMock:  mocks.NewMaterialServiceMock(mc),
		HistoryMock:   mocks.NewHistoryServiceMock(mc),
		EnumMock:      mocks.NewEnumServiceMock(mc),
		ParameterMock: mocks.NewParametersServiceMock(mc),
		AggregateMock: mocks.NewAggregateServiceMock(mc),
	}

	var (
		materials handlers.MaterialService       = env.MaterialMock
		history   handlers.HistoryService        = env.HistoryMock
		derived   handlers.DerivedProductService = env.DerivedMock
	)
	if opts.withoutMaterials {
		materials = nil
	}
	if opts.withoutHistory {
		history = nil
	}
	if opts.withoutDerived {
		derived = nil
	}

	env.Router = setupRouter(t, handlers.Handlers{
		ClassHandler:      handlers.NewClassHandlers(env.ClassMock),
		ProductHandler:    handlers.NewProductHandlers(env.ProductMock, derived),
		MaterialHandler:   handlers.NewMaterialHandlers(materials),
		HistoryHandler:    handlers.NewHistoryHandlers(history),
		EnumHandler:       handlers.NewEnumHandlers(env.EnumMock),
		ParametersHandler: handlers.NewParametersHandlers(env.ParameterMock),
		AggregatesHandler: handlers.NewAggregateHandlers(env.AggregateMock),
		AuthHandler:       handlers.NewAuthHandlers(),
	})

	resp := env.DoRequest(t.Context(), "", "POST", "/api/login", strings.NewReader(`{"username":"admin","password":"admin"}`))
	require.Equal(t, 200, resp.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	env.JWT = body["token"]

	return env
}

func (e *testEnv) DoRequest(ctx context.Context, jwt, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(ctx, method, path, body)
	req.Header.Set("Content-Type", "application/json")

	if jwt != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	}

	w := httptest.NewRecorder()
	e.Router.ServeHTTP(w, req)
	return w
}

// errorBody разбирает ответ в формате ErrorResponse.
func errorBody(t *testing.T, resp *httptest.ResponseRecorder) map[string]string {
	t.Helper()

	var got map[string]string
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &got))
	require.Contains(t, got, "error")
	require.Contains(t, got, "message")

	return got
}
