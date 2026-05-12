package handlers

import (
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// defaultOffset это дефолтное значение для смещения при пагинации.
	defaultOffset = 0

	// defaultLimit это дефолтное значения для лимита при пагинации.
	defaultLimit = 20

	// maxLimit это максимальное значение для лимита при пагинации.
	maxLimit = 100

	// filterParamArgsNum - кол-во аргументов параметра фильтрации.
	// Т. е. field:operation:number - 3 части.
	// Необходим для валидации параметра фильтрации
	filterParamArgsNum = 3
)

func writeError(c *gin.Context, code int, errType, message string) {
	c.JSON(code, gin.H{
		"error":   errType,
		"message": message,
	})
}

var numFields = []string{"rating", "abv", "ibu", "amount"}

func validateFilterParam(filter string) (map[string]any, error) {
	filterParams := strings.Split(filter, ":")
	if len(filterParams) != filterParamArgsNum {
		return nil, errors.New("неправильный параметр")
	}

	if !slices.Contains(numFields, filterParams[0]) {
		return nil, errors.New("неверное поле")
	}
	filterMap := map[string]any{
		"field": filterParams[0],
		"value": filterParams[2],
	}

	rawVal := filterParams[2]
	valInt, err := strconv.Atoi(rawVal)
	if err != nil {
		valFloat32, err := strconv.ParseFloat(rawVal, 32)
		if err != nil {
			return nil, errors.New(InvalidParameters)
		}
		filterMap["value"] = valFloat32
	} else {
		filterMap["value"] = valInt
	}

	switch filterParams[1] {
	case "eq", "gt", "ge", "lt", "le", "ne":
		filterMap["operation"] = filterParams[1]
	default:
		return nil, errors.New("неверная операция")
	}

	return filterMap, nil
}

// getUintParam извлекает и валидирует целочисленный ненулевой параметр из URL, например id.
func getUintParam(c *gin.Context, name string) (uint, error) {
	param := c.Param(name)
	if param == "" {
		return 0, nil
	}

	uintParam, err := strconv.Atoi(param)
	if err != nil || uintParam <= 0 {
		return 0, errors.New("invalid id")
	}

	return uint(uintParam), nil
}

// readRequestBody читает тело HTTP-запроса и возвращает его в виде байтового среза.
func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.New("invalid request body")
	}

	return body, nil
}

// getPaginationParams извлекает параметры пагинации из HTTP-запроса
func getPaginationParams(c *gin.Context) (uint64, uint64, error) {
	offset := defaultOffset
	limit := defaultLimit

	if rawOffset := c.Query("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil || parsedOffset < 0 {
			return 0, 0, errors.New("invalid offset")
		}

		offset = parsedOffset
	}

	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			return 0, 0, errors.New("invalid limit")
		}

		if parsedLimit > maxLimit {
			parsedLimit = maxLimit
		}

		limit = parsedLimit
	}

	return uint64(offset), uint64(limit), nil
}
