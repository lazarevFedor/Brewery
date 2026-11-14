// Package handlers содержит реализацию HTTP-обработчиков для каталога изделий.
package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"Brewery/pkg/logger"
	"errors"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// defaultOffset это дефолтное значение для смещения при пагинации.
	defaultOffset = 0

	// defaultLimit это дефолтное значения для лимита при пагинации.
	defaultLimit = 10

	// maxLimit это максимальное значение для лимита при пагинации.
	maxLimit = 100

	// filterParamArgsNum - кол-во аргументов параметра фильтрации.
	// Т. е. field:operation:value - 3 части.
	filterParamArgsNum = 3
)

// fieldNameRe ограничивает имя поля в фильтре, чтобы в запрос к БД не попало произвольное выражение.
var fieldNameRe = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// filterOperations сопоставляет операции из query-параметра с операциями сущности.
// Поддерживаются оба варианта записи: gte/lte (как в swagger) и ge/le.
var filterOperations = map[string]entities.FilterOperation{
	"eq":  entities.OpEqual,
	"ne":  entities.OpNotEqual,
	"gt":  entities.OpGreater,
	"ge":  entities.OpGreaterEqual,
	"gte": entities.OpGreaterEqual,
	"lt":  entities.OpLess,
	"le":  entities.OpLessEqual,
	"lte": entities.OpLessEqual,
}

// writeError отправляет ошибку в формате ErrorResponse из swagger.
func writeError(c *gin.Context, code int, errType, message string) {
	c.AbortWithStatusJSON(code, gin.H{
		"error":   errType,
		"message": message,
	})
}

// getLogger достает логгер из контекста запроса. Если логгера нет - сразу отвечает 500.
func getLogger(c *gin.Context) (*logger.Logger, bool) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
	}

	return log, ok
}

// parseFilter разбирает параметр фильтрации вида field:op:value.
func parseFilter(filter string) (entities.FilterParameter, error) {
	parts := strings.Split(filter, ":")
	if len(parts) != filterParamArgsNum {
		return entities.FilterParameter{}, errors.New("filter must be in the form field:op:value")
	}

	field, rawOp, rawVal := parts[0], parts[1], parts[2]
	if !fieldNameRe.MatchString(field) {
		return entities.FilterParameter{}, errors.New("invalid filter field")
	}

	op, ok := filterOperations[rawOp]
	if !ok {
		return entities.FilterParameter{}, errors.New("unsupported filter operation")
	}

	val, err := strconv.ParseFloat(rawVal, 32)
	if err != nil {
		return entities.FilterParameter{}, errors.New("filter value must be a number")
	}

	return entities.FilterParameter{
		FieldName: field,
		Operation: op,
		Value:     float32(val),
	}, nil
}

// getFilters разбирает все параметры filter из query. allowedOps ограничивает допустимые операции, nil - все.
func getFilters(c *gin.Context, allowedOps []entities.FilterOperation) ([]entities.FilterParameter, error) {
	rawFilters := c.QueryArray("filter")
	filters := make([]entities.FilterParameter, 0, len(rawFilters))

	for _, raw := range rawFilters {
		filter, err := parseFilter(raw)
		if err != nil {
			return nil, err
		}

		if allowedOps != nil && !slices.Contains(allowedOps, filter.Operation) {
			return nil, errors.New("filter operation is not allowed for this request")
		}

		filters = append(filters, filter)
	}

	return filters, nil
}

// getIDParam извлекает и валидирует положительный целочисленный параметр пути, например id.
func getIDParam(c *gin.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid " + name)
	}

	return id, nil
}

// toUint приводит id к uint для сервисов, которые принимают uint.
// Переполнения нет: id уже проверены getIDParam/getIDQuery и неотрицательны.
func toUint(id int64) uint {
	return uint(id) //nolint:gosec // id неотрицательный, проверен при разборе запроса
}

// getIDQuery извлекает положительный целочисленный query-параметр. Если необязательный параметр не передан, возвращает 0.
func getIDQuery(c *gin.Context, name string, required bool) (int64, error) {
	raw := c.Query(name)
	if raw == "" {
		if required {
			return 0, errors.New(name + " is required")
		}

		return 0, nil
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid " + name)
	}

	return id, nil
}

// getPaginationParams извлекает параметры пагинации из HTTP-запроса.
func getPaginationParams(c *gin.Context) (offset, limit int32, err error) {
	offset, limit = defaultOffset, defaultLimit

	if rawOffset := c.Query("offset"); rawOffset != "" {
		parsed, parseErr := strconv.ParseInt(rawOffset, 10, 32)
		if parseErr != nil || parsed < 0 {
			return 0, 0, errors.New("invalid offset")
		}

		offset = int32(parsed)
	}

	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, parseErr := strconv.ParseInt(rawLimit, 10, 32)
		if parseErr != nil || parsed <= 0 {
			return 0, 0, errors.New("invalid limit")
		}

		limit = int32(min(parsed, maxLimit))
	}

	return offset, limit, nil
}

// getPositiveFloatQuery извлекает необязательный положительный дробный query-параметр. Если он не передан, возвращает def.
func getPositiveFloatQuery(c *gin.Context, name string, def float32) (float32, error) {
	raw := c.Query(name)
	if raw == "" {
		return def, nil
	}

	val, err := strconv.ParseFloat(raw, 32)
	if err != nil || val <= 0 {
		return 0, errors.New(name + " must be a positive number")
	}

	return float32(val), nil
}

// getTimeQuery извлекает необязательный query-параметр с датой в формате RFC 3339.
func getTimeQuery(c *gin.Context, name string) (time.Time, error) {
	raw := c.Query(name)
	if raw == "" {
		return time.Time{}, nil
	}

	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, errors.New("invalid " + name + ", expected RFC 3339 date-time")
	}

	return t, nil
}

// bindJSON декодирует тело запроса в dst и при ошибке отвечает 400.
func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		writeError(c, http.StatusBadRequest, apperrors.CodeInvalidJSON, "Request body is not valid JSON")
		return false
	}

	return true
}

// bindUpdates декодирует тело PATCH-запроса в map и проверяет, что оно не пустое.
func bindUpdates(c *gin.Context) (map[string]any, bool) {
	updates := make(map[string]any)
	if !bindJSON(c, &updates) {
		return nil, false
	}

	if len(updates) == 0 {
		writeError(c, http.StatusBadRequest, apperrors.CodeBadRequest, "Empty update payload")
		return nil, false
	}

	return updates, true
}

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		panic("JWT_SECRET environment variable is not set")
	}
	return []byte(s)
}

func adminLogin() (string, string) {
	user := os.Getenv("ADMIN_USERNAME")
	pass := os.Getenv("ADMIN_PASSWORD")
	if user == "" {
		panic("ADMIN_USERNAME env variable is not set")
	}
	if pass == "" {
		panic("ADMIN_PASSWORD env variable is not set")
	}
	return user, pass
}
