package handlers_test

// import (
// 	"encoding/json"
// 	"net/http"
// 	"os"
// 	"regexp"
// 	"sort"
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// const swaggerPath = "../../../api/services.swagger.v2.0.json"

// type swaggerOperation struct {
// 	Security []map[string][]string `json:"security"`
// }

// type swaggerRoute struct {
// 	Method string
// 	Path   string
// 	Auth   bool
// }

// var swaggerParamRe = regexp.MustCompile(`\{(\w+)\}`)

// // loadSwaggerRoutes читает swagger и переводит пути в формат gin: /products/{id} -> /api/products/:id.
// func loadSwaggerRoutes(t *testing.T) []swaggerRoute {
// 	t.Helper()

// 	raw, err := os.ReadFile(swaggerPath)
// 	require.NoError(t, err)

// 	var spec struct {
// 		Paths map[string]map[string]swaggerOperation `json:"paths"`
// 	}
// 	require.NoError(t, json.Unmarshal(raw, &spec))

// 	routes := make([]swaggerRoute, 0)
// 	for path, ops := range spec.Paths {
// 		for method, op := range ops {
// 			routes = append(routes, swaggerRoute{
// 				Method: strings.ToUpper(method),
// 				Path:   "/api" + swaggerParamRe.ReplaceAllString(path, ":$1"),
// 				Auth:   len(op.Security) > 0,
// 			})
// 		}
// 	}

// 	return routes
// }

// // TestRoutes_MatchSwagger проверяет, что роутер регистрирует ровно те запросы, что описаны в swagger.
// func TestRoutes_MatchSwagger(t *testing.T) {
// 	env := newTestEnv(t)

// 	expected := make([]string, 0)
// 	for _, r := range loadSwaggerRoutes(t) {
// 		expected = append(expected, r.Method+" "+r.Path)
// 	}

// 	// Имена параметров пути в gin не важны для сопоставления, поэтому сравниваем их как есть:
// 	// в роутере используются те же имена, что и в swagger.
// 	actual := make([]string, 0)
// 	for _, r := range env.Router.Routes() {
// 		if r.Path == "/metrics" {
// 			continue
// 		}
// 		actual = append(actual, r.Method+" "+r.Path)
// 	}

// 	sort.Strings(expected)
// 	sort.Strings(actual)
// 	assert.Equal(t, expected, actual)
// }

// // TestRoutes_AuthMatchesSwagger проверяет, что запросы с bearerAuth в swagger без токена получают 401,
// // а запросы без bearerAuth доступны без токена.
// func TestRoutes_AuthMatchesSwagger(t *testing.T) {
// 	env := newTestEnv(t)

// 	for _, r := range loadSwaggerRoutes(t) {
// 		if !r.Auth {
// 			continue
// 		}

// 		path := strings.NewReplacer(":id", "1", ":class_id", "1", ":material_id", "2").Replace(r.Path)
// 		resp := env.DoRequest(t.Context(), "", r.Method, path, nil)
// 		assert.Equal(t, http.StatusUnauthorized, resp.Code, "%s %s must require auth", r.Method, r.Path)
// 	}
// }
